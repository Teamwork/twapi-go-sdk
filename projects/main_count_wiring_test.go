package projects_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// countWiring records, per list endpoint, the four facts that have to agree for
// a total count to be reported honestly.
type countWiring struct {
	sharedMeta bool // the response's Meta is twapi.ListMeta
	countMode  bool // the filters carry a CountMode twapi.ListCountMode slot
	applied    bool // the filters' apply calls CountMode.Apply
	setRequest bool // the response has a SetRequest at all
	resolved   bool // that SetRequest calls Meta.ResolveCount
}

// TestCountWiring asserts that every list response carrying twapi.ListMeta has
// the count wiring described in AGENTS.md. Unlike the sparse-fieldsets wiring,
// nothing generates this and the compiler only catches part of it: a brand-new
// list response that uses twapi.ListMeta without a CountMode slot builds
// cleanly, and a SetRequest that forgets Meta.ResolveCount silently surfaces
// the API's lower-bound count as if it were a total.
func TestCountWiring(t *testing.T) {
	endpoints, err := collectCountWiring(".")
	if err != nil {
		t.Fatalf("failed to parse package: %s", err)
	}

	var checked int
	for _, base := range slices.Sorted(maps.Keys(endpoints)) {
		w := endpoints[base]
		if !w.sharedMeta {
			continue
		}
		// a response with no SetRequest has nowhere to reconcile the count, so it
		// must not offer the option at all
		if !w.setRequest {
			if w.countMode {
				t.Errorf("%s: offers CountMode but has no SetRequest to resolve the count", base)
			}
			continue
		}
		checked++
		if !w.countMode {
			t.Errorf("%s: response uses twapi.ListMeta but its filters have no CountMode slot", base)
		}
		if !w.applied {
			t.Errorf("%s: filters carry CountMode but apply never calls CountMode.Apply", base)
		}
		if !w.resolved {
			t.Errorf("%s: SetRequest never calls Meta.ResolveCount, so a skipped count would "+
				"reach callers as a total", base)
		}
	}

	if checked == 0 {
		t.Fatal("found no wired list responses, so this test is asserting nothing")
	}
	t.Logf("verified count wiring on %d list responses", checked)
}

// collectCountWiring parses every non-test file in dir and indexes what it
// finds by endpoint base name, e.g. "MessageList" for MessageListResponse and
// MessageListRequestFilters.
func collectCountWiring(dir string) (map[string]*countWiring, error) {
	found := map[string]*countWiring{}
	at := func(base string) *countWiring {
		if found[base] == nil {
			found[base] = &countWiring{}
		}
		return found[base]
	}

	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.TypeSpec:
				indexStructFields(node, at)
			case *ast.FuncDecl:
				indexMethodBody(node, at)
			}
			return true
		})
	}
	return found, nil
}

// indexStructFields records the Meta and CountMode slots of a struct.
func indexStructFields(node *ast.TypeSpec, at func(string) *countWiring) {
	st, ok := node.Type.(*ast.StructType)
	if !ok {
		return
	}
	for _, f := range st.Fields.List {
		sel, ok := f.Type.(*ast.SelectorExpr)
		if !ok || len(f.Names) != 1 {
			continue
		}
		field, typ, name := f.Names[0].Name, sel.Sel.Name, node.Name.Name
		switch {
		case field == "Meta" && typ == "ListMeta" && strings.HasSuffix(name, "Response"):
			at(strings.TrimSuffix(name, "Response")).sharedMeta = true
		case field == "CountMode" && typ == "ListCountMode" && strings.HasSuffix(name, "RequestFilters"):
			at(strings.TrimSuffix(name, "RequestFilters")).countMode = true
		}
	}
}

// indexMethodBody records whether a filter's apply and a response's SetRequest
// make the calls the wiring depends on.
func indexMethodBody(node *ast.FuncDecl, at func(string) *countWiring) {
	if node.Recv == nil || len(node.Recv.List) != 1 || node.Body == nil {
		return
	}
	recv := receiverTypeName(node.Recv.List[0].Type)
	switch {
	case node.Name.Name == "apply" && strings.HasSuffix(recv, "RequestFilters"):
		base := strings.TrimSuffix(recv, "RequestFilters")
		at(base).applied = callsFieldMethod(node.Body, "CountMode", "Apply")
	case node.Name.Name == "SetRequest" && strings.HasSuffix(recv, "Response"):
		base := strings.TrimSuffix(recv, "Response")
		at(base).setRequest = true
		at(base).resolved = callsFieldMethod(node.Body, "Meta", "ResolveCount")
	}
}

func receiverTypeName(e ast.Expr) string {
	if star, ok := e.(*ast.StarExpr); ok {
		e = star.X
	}
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

// callsFieldMethod reports whether body contains a `<recv>.<field>.<method>(…)`
// call, e.g. `m.CountMode.Apply(query)`.
func callsFieldMethod(body *ast.BlockStmt, field, method string) bool {
	var found bool
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		outer, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || outer.Sel.Name != method {
			return true
		}
		if inner, ok := outer.X.(*ast.SelectorExpr); ok && inner.Sel.Name == field {
			found = true
		}
		return true
	})
	return found
}
