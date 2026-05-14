// Command sparsefieldsgen generates typed string enums and per-list containers
// for v3 sparse fieldsets.
//
// It scans a Go package for struct type declarations that carry either of two
// markers on their doc comment:
//
//   - sparsefields:gen [=CustomTypeName]
//     Marks an entity struct (e.g. Task, Project). Emits a named string type
//     plus one constant per JSON-tagged field on the struct (same-package
//     embedded structs are flattened).
//
//   - sparsefields:list [=CustomTypeName]
//     Marks a *ListResponse struct. Emits a typed container struct (e.g.
//     TaskListFields) whose fields mirror the response's main slice and each
//     entry of its Included sub-struct, plus an `apply(url.Values)` method
//     that writes the appropriate fields[entityKey]=… query parameters via
//     twapi.ApplySparseFields.
//
// Usage (typically invoked via //go:generate from the target package):
//
//	//go:generate go run github.com/teamwork/twapi-go-sdk/internal/sparsefieldsgen
//
// By default it scans the current working directory and writes
// `sparse_fields_gen.go` next to the sources. Both can be overridden with the
// -src and -out flags.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const (
	markerEntity = "sparsefields:gen"
	markerList   = "sparsefields:list"

	rootImportPath = "github.com/teamwork/twapi-go-sdk"
	rootImportName = "twapi"
)

type entity struct {
	StructName string
	FieldType  string
	Fields     []fieldEntry
}

type fieldEntry struct {
	GoName  string
	JSONTag string
}

type listResponse struct {
	StructName    string // e.g. "TaskListResponse"
	ContainerName string // e.g. "TaskListFields"
	Slots         []slot
}

type slot struct {
	GoName      string // field name on the container (e.g. "Tasks")
	EntityKey   string // JSON tag used in fields[...]= (e.g. "tasks")
	ElementType string // generated Field-type name (e.g. "TaskField")
}

func main() {
	src := flag.String("src", ".", "directory containing the package to scan")
	out := flag.String("out", "sparse_fields_gen.go", "output file name (written into -src)")
	outTest := flag.String("out-test", "sparse_fields_gen_test.go", "output file for generated tests (written into -src)")
	flag.Parse()

	abs, err := filepath.Abs(*src)
	if err != nil {
		log.Fatalf("sparsefieldsgen: resolve src: %v", err)
	}

	fset := token.NewFileSet()
	skipName := map[string]bool{*out: true, *outTest: true}
	files, pkgName, err := parsePackage(fset, abs, skipName)
	if err != nil {
		log.Fatalf("sparsefieldsgen: parse %s: %v", abs, err)
	}
	if len(files) == 0 {
		log.Fatalf("sparsefieldsgen: no Go package found in %s", abs)
	}

	var (
		entities []entity
		lists    []listResponse
		structs  = map[string]*ast.StructType{}
	)
	for _, file := range files {
		indexStructs(file, structs)
	}
	for _, file := range files {
		collectEntities(file, structs, &entities)
	}
	if len(entities) == 0 {
		log.Fatalf("sparsefieldsgen: no `%s` markers found in %s", markerEntity, abs)
	}

	// Build entity → FieldType lookup, used by the list pass to resolve slot
	// element types (e.g. Task → TaskField).
	fieldTypeOf := map[string]string{}
	for _, e := range entities {
		fieldTypeOf[e.StructName] = e.FieldType
	}

	for _, file := range files {
		collectLists(file, fieldTypeOf, &lists)
	}

	sort.Slice(entities, func(i, j int) bool {
		return entities[i].FieldType < entities[j].FieldType
	})
	sort.Slice(lists, func(i, j int) bool {
		return lists[i].ContainerName < lists[j].ContainerName
	})

	source, err := render(pkgName, entities, lists)
	if err != nil {
		log.Fatalf("sparsefieldsgen: render: %v", err)
	}

	dest := filepath.Join(abs, *out)
	if err := os.WriteFile(dest, source, 0o644); err != nil {
		log.Fatalf("sparsefieldsgen: write %s: %v", dest, err)
	}

	// Generated tests cover the lists whose filter already exposes a
	// `Fields <Container>` slot. Lists in flight (marker added but filter not
	// yet wired) are skipped silently so regeneration doesn't break the build.
	entityByField := map[string]entity{}
	for _, e := range entities {
		entityByField[e.FieldType] = e
	}
	wiredLists := wiredLists(lists, structs)

	// Static wiring check: a filter declaring `Fields <Container>` must also
	// invoke `*.Fields.apply(...)` somewhere in one of its methods, otherwise
	// the response will silently ignore sparse-field selections. Failing at
	// generate time prevents a half-wired filter from shipping unnoticed.
	if missing := unwiredFilters(wiredLists, files); len(missing) > 0 {
		log.Fatalf("sparsefieldsgen: filters declare a Fields slot but never call `<receiver>.Fields.apply(...)`:\n  "+
			"- %s\n"+
			"Add `t.Fields.apply(query)` (or equivalent) to each filter's method that mutates the request.",
			strings.Join(missing, "\n  - "))
	}

	testDest := filepath.Join(abs, *outTest)
	if len(wiredLists) == 0 {
		// Remove any stale generated test file from a previous run so the file
		// doesn't outlive the markers that produced it.
		if err := os.Remove(testDest); err != nil && !os.IsNotExist(err) {
			log.Fatalf("sparsefieldsgen: remove stale %s: %v", testDest, err)
		}
		return
	}
	importPath, err := packageImportPath(abs)
	if err != nil {
		log.Fatalf("sparsefieldsgen: resolve import path: %v", err)
	}
	testSource, err := renderTests(pkgName, importPath, wiredLists, entityByField)
	if err != nil {
		log.Fatalf("sparsefieldsgen: render tests: %v", err)
	}
	if err := os.WriteFile(testDest, testSource, 0o644); err != nil {
		log.Fatalf("sparsefieldsgen: write %s: %v", testDest, err)
	}
}

// parsePackage reads dir, parses every non-test `.go` file (skipping the
// generator's own outputs in skipName), and returns the files belonging to the
// single non-test package found there plus that package's name. It replaces
// the deprecated parser.ParseDir while keeping the same filtering rules — we
// don't need build-tag-aware loading because the generator only inspects
// struct declarations and their doc comments.
func parsePackage(fset *token.FileSet, dir string, skipName map[string]bool) ([]*ast.File, string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", fmt.Errorf("read dir: %w", err)
	}

	var (
		files   []*ast.File
		pkgName string
	)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || skipName[name] {
			continue
		}
		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, "", fmt.Errorf("parse %s: %w", name, err)
		}
		filePkg := file.Name.Name
		if strings.HasSuffix(filePkg, "_test") {
			continue
		}
		if pkgName == "" {
			pkgName = filePkg
		} else if pkgName != filePkg {
			return nil, "", fmt.Errorf("multiple packages in %s: %s and %s", dir, pkgName, filePkg)
		}
		files = append(files, file)
	}
	return files, pkgName, nil
}

// wiredLists returns the subset of lists whose <Request>Filters struct already
// exposes a `Fields <Container>` field. Other lists are skipped so emitting a
// test wouldn't break compilation while the rollout is in progress.
func wiredLists(lists []listResponse, structs map[string]*ast.StructType) []listResponse {
	var out []listResponse
	for _, l := range lists {
		filtersName := strings.TrimSuffix(l.StructName, "Response") + "RequestFilters"
		st, ok := structs[filtersName]
		if !ok {
			continue
		}
		if hasField(st, "Fields", l.ContainerName) {
			out = append(out, l)
		}
	}
	return out
}

// unwiredFilters returns one human-readable diagnostic per wired list whose
// filter declares `Fields <Container>` but never invokes
// `<receiver>.Fields.apply(...)` from any of its own methods. An empty result
// means every wired filter actually pipes its sparse-fields selection through
// at request-build time.
func unwiredFilters(lists []listResponse, files []*ast.File) []string {
	var missing []string
	for _, l := range lists {
		filterName := strings.TrimSuffix(l.StructName, "Response") + "RequestFilters"
		if !hasFieldsApplyCall(files, filterName) {
			missing = append(missing, fmt.Sprintf("%s declares `Fields %s` but no method on %s calls `*.Fields.apply(...)`",
				filterName, l.ContainerName, filterName))
		}
	}
	return missing
}

// hasFieldsApplyCall reports whether any method whose receiver is recvName
// (value or pointer) contains a call shaped like `<x>.Fields.apply(<args>)`.
// The check is intentionally lenient about the receiver/argument identifiers
// so it accepts the conventional `t.Fields.apply(query)` and any equivalent.
func hasFieldsApplyCall(files []*ast.File, recvName string) bool {
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 || fn.Body == nil {
				continue
			}
			if recvTypeName(fn.Recv.List[0].Type) != recvName {
				continue
			}
			if containsFieldsApplyCall(fn.Body) {
				return true
			}
		}
	}
	return false
}

// recvTypeName returns the unqualified type name of a method receiver
// expression, peeling off a leading pointer if any.
func recvTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	}
	return ""
}

// containsFieldsApplyCall walks node and reports whether it contains a call
// expression whose function selector is `<x>.Fields.apply` — the canonical
// shape of the call we expect filters to make.
func containsFieldsApplyCall(node ast.Node) bool {
	var found bool
	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		outer, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || outer.Sel.Name != "apply" {
			return true
		}
		inner, ok := outer.X.(*ast.SelectorExpr)
		if !ok || inner.Sel.Name != "Fields" {
			return true
		}
		found = true
		return false
	})
	return found
}

// hasField reports whether st declares a field named name with the given
// unqualified type.
func hasField(st *ast.StructType, name, typeName string) bool {
	for _, f := range st.Fields.List {
		id, ok := f.Type.(*ast.Ident)
		if !ok || id.Name != typeName {
			continue
		}
		for _, n := range f.Names {
			if n.Name == name {
				return true
			}
		}
	}
	return false
}

// indexStructs records every named struct type defined in file. Used so that
// extractFields can flatten same-package embedded types.
func indexStructs(file *ast.File, dst map[string]*ast.StructType) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			dst[ts.Name.Name] = st
		}
	}
}

// collectEntities walks file and appends an entity for every struct that
// carries the sparsefields:gen marker.
func collectEntities(file *ast.File, structs map[string]*ast.StructType, dst *[]entity) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			doc := ts.Doc
			if doc == nil {
				doc = gen.Doc
			}
			fieldType, ok := markerOverride(doc, markerEntity, ts.Name.Name+"Field")
			if !ok {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				log.Fatalf("sparsefieldsgen: %s is marked %q but is not a struct", ts.Name.Name, markerEntity)
			}
			fields, err := extractFields(st, structs, map[string]bool{ts.Name.Name: true})
			if err != nil {
				log.Fatalf("sparsefieldsgen: %s: %v", ts.Name.Name, err)
			}
			if len(fields) == 0 {
				log.Fatalf("sparsefieldsgen: %s has no json-tagged fields", ts.Name.Name)
			}
			*dst = append(*dst, entity{
				StructName: ts.Name.Name,
				FieldType:  fieldType,
				Fields:     fields,
			})
		}
	}
}

// collectLists walks file and appends a listResponse for every struct that
// carries the sparsefields:list marker.
func collectLists(file *ast.File, fieldTypeOf map[string]string, dst *[]listResponse) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			doc := ts.Doc
			if doc == nil {
				doc = gen.Doc
			}
			container, ok := markerOverride(doc, markerList, defaultContainerName(ts.Name.Name))
			if !ok {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				log.Fatalf("sparsefieldsgen: %s is marked %q but is not a struct", ts.Name.Name, markerList)
			}
			slots, err := extractSlots(st, ts.Name.Name, fieldTypeOf)
			if err != nil {
				log.Fatalf("sparsefieldsgen: %s: %v", ts.Name.Name, err)
			}
			if len(slots) == 0 {
				log.Fatalf("sparsefieldsgen: %s has no sparse-fields slots (no slice field nor Included struct)", ts.Name.Name)
			}
			*dst = append(*dst, listResponse{
				StructName:    ts.Name.Name,
				ContainerName: container,
				Slots:         slots,
			})
		}
	}
}

// defaultContainerName drops a trailing "Response" if present so
// TaskListResponse → TaskListFields.
func defaultContainerName(structName string) string {
	trimmed := strings.TrimSuffix(structName, "Response")
	return trimmed + "Fields"
}

// markerOverride returns the requested name and whether the marker is present
// on doc. Supports `marker` (use defaultName) and `marker=Name` forms.
func markerOverride(doc *ast.CommentGroup, marker, defaultName string) (string, bool) {
	if doc == nil {
		return "", false
	}
	for _, c := range doc.List {
		text := strings.TrimPrefix(c.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if !strings.HasPrefix(text, marker) {
			continue
		}
		rest := strings.TrimPrefix(text, marker)
		switch {
		case rest == "":
			return defaultName, true
		case strings.HasPrefix(rest, "="):
			name := strings.TrimSpace(strings.TrimPrefix(rest, "="))
			if name == "" {
				return defaultName, true
			}
			return name, true
		}
	}
	return "", false
}

// extractFields returns the json-tagged fields of a struct, in source order.
// Fields tagged `json:"-"` and fields without a json tag are skipped. Embedded
// types defined in the same package are flattened recursively. When an outer
// field and an embedded field share the same JSON tag (Go's regular shadowing
// rules), the outer field wins.
func extractFields(
	st *ast.StructType,
	structs map[string]*ast.StructType,
	visited map[string]bool,
) ([]fieldEntry, error) {
	var out []fieldEntry
	seen := map[string]int{} // jsonTag -> index in out
	add := func(entry fieldEntry) {
		if idx, ok := seen[entry.JSONTag]; ok {
			out[idx] = entry
			return
		}
		seen[entry.JSONTag] = len(out)
		out = append(out, entry)
	}
	// First pass: add the struct's own (non-embedded) fields so they shadow
	// anything contributed by embedded types.
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 || field.Tag == nil {
			continue
		}
		tagText, err := strconv.Unquote(field.Tag.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid struct tag %s: %w", field.Tag.Value, err)
		}
		jsonTag := reflect.StructTag(tagText).Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		name := strings.SplitN(jsonTag, ",", 2)[0]
		if name == "" {
			continue
		}
		for _, ident := range field.Names {
			if !ident.IsExported() {
				continue
			}
			add(fieldEntry{
				GoName:  ident.Name,
				JSONTag: name,
			})
		}
	}
	// Second pass: flatten embedded types, but never overwrite a name already
	// contributed by the outer struct.
	for _, field := range st.Fields.List {
		if len(field.Names) != 0 || field.Tag != nil {
			continue
		}
		ident, ok := embeddedIdent(field.Type)
		if !ok {
			continue
		}
		embedded, ok := structs[ident.Name]
		if !ok || visited[ident.Name] {
			continue
		}
		visited[ident.Name] = true
		nested, err := extractFields(embedded, structs, visited)
		if err != nil {
			return nil, err
		}
		for _, entry := range nested {
			if _, ok := seen[entry.JSONTag]; ok {
				continue
			}
			add(entry)
		}
	}
	return out, nil
}

// extractSlots inspects a *ListResponse struct and returns one slot per
// sparse-fields target it exposes:
//
//   - the (single) top-level slice field whose element type has an entity Field
//     enum becomes the main slot;
//   - each map field inside the Included sub-struct whose value type has an
//     entity Field enum becomes a sideload slot.
//
// Fields whose element type isn't in fieldTypeOf are reported as errors so the
// generator can't silently drop a slot.
func extractSlots(st *ast.StructType, ownerName string, fieldTypeOf map[string]string) ([]slot, error) {
	var slots []slot
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 || field.Tag == nil {
			continue
		}
		tagText, err := strconv.Unquote(field.Tag.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid struct tag %s: %w", field.Tag.Value, err)
		}
		jsonTag := strings.SplitN(reflect.StructTag(tagText).Get("json"), ",", 2)[0]
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		switch t := field.Type.(type) {
		case *ast.ArrayType:
			elem, ok := elementIdent(t.Elt)
			if !ok {
				continue
			}
			fieldType, ok := fieldTypeOf[elem.Name]
			if !ok {
				return nil, fmt.Errorf("slice field %s.%s uses element type %s which has no `%s` marker",
					ownerName, field.Names[0].Name, elem.Name, markerEntity)
			}
			slots = append(slots, slot{
				GoName:      field.Names[0].Name,
				EntityKey:   jsonTag,
				ElementType: fieldType,
			})
		case *ast.StructType:
			// Anonymous struct field. Only the conventional "Included" container
			// is walked; "Meta" and similar bookkeeping fields are ignored.
			if field.Names[0].Name != "Included" {
				continue
			}
			sub, err := extractIncludedSlots(t, ownerName, fieldTypeOf)
			if err != nil {
				return nil, err
			}
			slots = append(slots, sub...)
		}
	}
	return slots, nil
}

// extractIncludedSlots walks an anonymous Included struct, emitting one slot
// per map<…, EntityType> field whose value type has an entity Field enum.
func extractIncludedSlots(st *ast.StructType, ownerName string, fieldTypeOf map[string]string) ([]slot, error) {
	var slots []slot
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 || field.Tag == nil {
			continue
		}
		mt, ok := field.Type.(*ast.MapType)
		if !ok {
			continue
		}
		elem, ok := elementIdent(mt.Value)
		if !ok {
			continue
		}
		fieldType, ok := fieldTypeOf[elem.Name]
		if !ok {
			return nil, fmt.Errorf("sideload %s.Included.%s uses value type %s which has no `%s` marker",
				ownerName, field.Names[0].Name, elem.Name, markerEntity)
		}
		tagText, err := strconv.Unquote(field.Tag.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid struct tag %s: %w", field.Tag.Value, err)
		}
		jsonTag := strings.SplitN(reflect.StructTag(tagText).Get("json"), ",", 2)[0]
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		slots = append(slots, slot{
			GoName:      field.Names[0].Name,
			EntityKey:   jsonTag,
			ElementType: fieldType,
		})
	}
	return slots, nil
}

// elementIdent returns an unqualified ident for a same-package type reference,
// stripping pointer wrappers. Returns false for qualified types (e.g.
// twapi.Relationship) and structural types (slice-of-slice, etc.).
func elementIdent(expr ast.Expr) (*ast.Ident, bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t, true
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id, true
		}
	}
	return nil, false
}

// embeddedIdent unwraps a pointer-or-bare embedded type expression to its
// identifier, when that identifier is unqualified.
func embeddedIdent(expr ast.Expr) (*ast.Ident, bool) {
	return elementIdent(expr)
}

func render(pkgName string, entities []entity, lists []listResponse) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "// Code generated by sparsefieldsgen; DO NOT EDIT.")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "package %s\n\n", pkgName)

	if len(lists) > 0 {
		fmt.Fprintln(&buf, "import (")
		fmt.Fprintln(&buf, "\t\"net/url\"")
		fmt.Fprintln(&buf)
		fmt.Fprintf(&buf, "\t%s %q\n", rootImportName, rootImportPath)
		fmt.Fprintln(&buf, ")")
		fmt.Fprintln(&buf)
	}

	for i, e := range entities {
		if i > 0 {
			fmt.Fprintln(&buf)
		}
		fmt.Fprintf(&buf, "// %s identifies a JSON-tagged attribute of %s usable for v3 sparse fieldsets.\n",
			e.FieldType, e.StructName)
		fmt.Fprintf(&buf, "type %s string\n\n", e.FieldType)
		fmt.Fprintf(&buf, "// List of possible %s fields.\n", e.StructName)
		fmt.Fprintln(&buf, "const (")
		for _, f := range e.Fields {
			constName := e.FieldType + f.GoName
			fmt.Fprintf(&buf, "\t%s %s = %q\n", constName, e.FieldType, f.JSONTag)
		}
		fmt.Fprintln(&buf, ")")
	}

	for _, l := range lists {
		fmt.Fprintln(&buf)
		fmt.Fprintf(&buf, "// %s selects sparse-fields slots for %s. Leave a slot empty to receive the\n",
			l.ContainerName, l.StructName)
		fmt.Fprintln(&buf, "// API default for that entity; populate it to restrict the attributes returned.")
		fmt.Fprintf(&buf, "type %s struct {\n", l.ContainerName)
		for _, s := range l.Slots {
			fmt.Fprintf(&buf, "\t// %s controls fields[%s]=… on the response.\n", s.GoName, s.EntityKey)
			fmt.Fprintf(&buf, "\t%s []%s\n", s.GoName, s.ElementType)
		}
		fmt.Fprintln(&buf, "}")
		fmt.Fprintln(&buf)
		fmt.Fprintf(&buf, "// apply writes every populated slot to query as a fields[entity]=… parameter.\n")
		fmt.Fprintf(&buf, "func (f %s) apply(query url.Values) {\n", l.ContainerName)
		for _, s := range l.Slots {
			fmt.Fprintf(&buf, "\t%s.ApplySparseFields(query, %q, f.%s)\n", rootImportName, s.EntityKey, s.GoName)
		}
		fmt.Fprintln(&buf, "}")
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), fmt.Errorf("format generated source: %w", err)
	}
	return formatted, nil
}

// renderTests emits a *_test.go file in the same package as the source
// (internal tests) containing, for each wired list, two tests:
//   - <Container>Apply: populates every slot with its entity's first constant
//     and asserts the resulting fields[entity]=… query parameters.
//   - <Container>ZeroValue: ensures a zero-value container emits no fields[*]=…
//     parameters, guarding against accidental wiring that always writes them.
//
// The container's apply method is called directly so the test works for every
// list whether or not its enclosing request type requires path arguments to
// build an HTTP request.
func renderTests(pkgName, _ string, lists []listResponse, entityByField map[string]entity) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "// Code generated by sparsefieldsgen; DO NOT EDIT.")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "package %s\n\n", pkgName)

	fmt.Fprintln(&buf, "import (")
	fmt.Fprintln(&buf, "\t\"net/url\"")
	fmt.Fprintln(&buf, "\t\"strings\"")
	fmt.Fprintln(&buf, "\t\"testing\"")
	fmt.Fprintln(&buf, ")")

	for _, l := range lists {
		// Resolve a representative (first) constant per slot, plus its JSON value.
		type slotConst struct {
			GoName    string
			Element   string
			ConstName string
			JSONValue string
			EntityKey string
		}
		consts := make([]slotConst, 0, len(l.Slots))
		for _, s := range l.Slots {
			e, ok := entityByField[s.ElementType]
			if !ok || len(e.Fields) == 0 {
				return nil, fmt.Errorf("%s: slot %s references unknown field type %s",
					l.StructName, s.GoName, s.ElementType)
			}
			first := e.Fields[0]
			consts = append(consts, slotConst{
				GoName:    s.GoName,
				Element:   s.ElementType,
				ConstName: e.FieldType + first.GoName,
				JSONValue: first.JSONTag,
				EntityKey: s.EntityKey,
			})
		}

		fmt.Fprintln(&buf)
		fmt.Fprintf(&buf, "// Test%sApply verifies that populated %s slots emit the\n",
			l.ContainerName, l.ContainerName)
		fmt.Fprintln(&buf, "// expected fields[entity]=… query parameters.")
		fmt.Fprintf(&buf, "func Test%sApply(t *testing.T) {\n", l.ContainerName)
		fmt.Fprintf(&buf, "\tfields := %s{\n", l.ContainerName)
		for _, c := range consts {
			fmt.Fprintf(&buf, "\t\t%s: []%s{%s},\n", c.GoName, c.Element, c.ConstName)
		}
		fmt.Fprintln(&buf, "\t}")
		buf.WriteString("\tquery := url.Values{}\n")
		buf.WriteString("\tfields.apply(query)\n")
		buf.WriteString("\tchecks := map[string]string{\n")
		for _, c := range consts {
			fmt.Fprintf(&buf, "\t\t\"fields[%s]\": %q,\n", c.EntityKey, c.JSONValue)
		}
		buf.WriteString("\t}\n")
		buf.WriteString("\tfor key, want := range checks {\n")
		buf.WriteString("\t\tif got := query.Get(key); got != want {\n")
		buf.WriteString("\t\t\tt.Errorf(\"%s = %q, want %q\", key, got, want)\n")
		buf.WriteString("\t\t}\n")
		buf.WriteString("\t}\n")
		buf.WriteString("}\n")

		fmt.Fprintln(&buf)
		fmt.Fprintf(&buf, "// Test%sZeroValue verifies that an unset %s emits no\n",
			l.ContainerName, l.ContainerName)
		buf.WriteString("// fields[*]=… query parameters.\n")
		fmt.Fprintf(&buf, "func Test%sZeroValue(t *testing.T) {\n", l.ContainerName)
		fmt.Fprintf(&buf, "\tvar fields %s\n", l.ContainerName)
		buf.WriteString("\tquery := url.Values{}\n")
		buf.WriteString("\tfields.apply(query)\n")
		buf.WriteString("\tfor key := range query {\n")
		buf.WriteString("\t\tif strings.HasPrefix(key, \"fields[\") {\n")
		buf.WriteString("\t\t\tt.Errorf(\"unexpected sparse-fields parameter %q on zero-value container\", key)\n")
		buf.WriteString("\t\t}\n")
		buf.WriteString("\t}\n")
		buf.WriteString("}\n")
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), fmt.Errorf("format generated tests: %w", err)
	}
	return formatted, nil
}

// packageImportPath returns the import path of the package rooted at dir,
// asking `go list` to resolve it so the result respects modules, replace
// directives, and the workspace configuration.
func packageImportPath(dir string) (string, error) {
	cmd := exec.Command("go", "list", "-f", "{{.ImportPath}}")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go list (in %s): %w", dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}
