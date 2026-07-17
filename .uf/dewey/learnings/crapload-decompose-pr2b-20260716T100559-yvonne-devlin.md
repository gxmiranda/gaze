---
tag: crapload-decompose-pr2b
author: yvonne-devlin
category: pattern
created_at: 2026-07-16T10:05:59Z
identity: crapload-decompose-pr2b-20260716T100559-yvonne-devlin
tier: draft
---

The parseAndTypeCheck test helper pattern in internal/quality/container_unwrap_internal_test.go is highly effective for testing Go AST manipulation functions. It takes a Go source string, parses it, and type-checks it with Importer:nil (intentionally ignoring import resolution errors). This enables testing functions that operate on types.Object, types.Info, and ast.File without loading real packages. For functions that also need *packages.Package (like traceForwardDataFlow which needs Syntax for file iteration), construct a minimal struct: packages.Package{Syntax: []*ast.File{file}, TypesInfo: info, Fset: fset}. The parseAndTypeCheckWithFset variant was added in Phase 2b to return the fset needed for this pattern. One gotcha: for isByteLikeParam and isPointerDestParam tests, it's simpler to construct types.Type values directly using the go/types constructors (types.NewSlice, types.NewPointer, types.NewInterfaceType) rather than parsing Go source — the classifiers only need type metadata, not full AST context.
