package projects

//go:generate go run github.com/teamwork/twapi-go-sdk/internal/sparsefieldsgen

// Sparse fieldsets are described at
// https://apidocs.teamwork.com/guides/teamwork/sparse-fieldsets
//
// Per-entity field enums and per-list `<List>Fields` containers are emitted
// by `sparsefieldsgen` into sparse_fields_gen.go. The generic helper used by
// the generated `apply` methods lives at the module root as
// `twapi.ApplySparseFields`.
