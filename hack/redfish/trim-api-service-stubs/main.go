// Command trim-api-service-stubs removes every pkg/redfish/api_service.go
// APIService method that still returns the generated "not implemented"
// stub body, keeping only the operations pkg/redfish actually implements.
//
// api_service.go was generated once, then moved out of
// pkg/generated/redfish/server so it could be hand-implemented; it has
// never been regenerated since, and hack/redfish/generate.sh doesn't touch
// it. Before hack/redfish/trim-redfish-spec existed, that was harmless --
// the file simply carried a stub for every operation the untrimmed DMTF
// spec defined. Now that the spec is trimmed to the implemented operation
// set before generation, most of those stubs reference model types that no
// longer exist in pkg/generated/redfish/server, and the file won't compile
// until they're gone.
//
// This is a one-time migration, not a step in the regular generate.sh
// pipeline: run it once to reduce api_service.go to the real
// implementations. From then on, adding a new endpoint means writing its
// method by hand against the interface in
// pkg/generated/redfish/server/api.go, the same way the methods that
// remain were written -- there's no longer a pre-existing stub to fill in.
//
// Usage:
//
//	go run ./hack/redfish/trim-api-service-stubs -file pkg/redfish/api_service.go
//	goimports -w pkg/redfish/api_service.go
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
)

func main() {
	path := flag.String("file", "pkg/redfish/api_service.go", "hand-maintained service file to trim in place")
	flag.Parse()

	if err := run(*path); err != nil {
		fmt.Fprintf(os.Stderr, "trim-api-service-stubs: %v\n", err)
		os.Exit(1)
	}
}

func run(path string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	removed := trimStubs(f)

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer out.Close() //nolint:errcheck

	if err := writeFile(out, fset, f); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	fmt.Fprintf(os.Stderr, "trim-api-service-stubs: removed %d unimplemented stub(s) -> %s\n", removed, path)
	return nil
}

// span is a half-open byte-position range within the source file.
type span struct {
	start, end token.Pos
}

func (s span) contains(cg *ast.CommentGroup) bool {
	return cg.Pos() >= s.start && cg.End() <= s.end
}

// trimStubs drops every APIService method in f that's still a generated
// "not implemented" stub, and reports how many it removed. Each generated
// stub carries several free-standing "// TODO: ..." comments inside its
// body, in addition to its doc comment -- deleting only the FuncDecl leaves
// those comments in f.Comments with no surviving node at their position, so
// go/printer prints them floating wherever their byte offset now falls.
// Removing every comment whose span lies within a deleted declaration
// avoids that.
func trimStubs(f *ast.File) int {
	var removedDecls int
	var removedSpans []span

	kept := f.Decls[:0]
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && isAPIServiceMethod(fn) && isNotImplementedStub(fn) {
			removedDecls++
			start := decl.Pos()
			if fn.Doc != nil {
				start = fn.Doc.Pos()
			}
			removedSpans = append(removedSpans, span{start, decl.End()})
			continue
		}
		kept = append(kept, decl)
	}
	f.Decls = kept

	comments := f.Comments[:0]
	for _, cg := range f.Comments {
		orphaned := false
		for _, s := range removedSpans {
			if s.contains(cg) {
				orphaned = true
				break
			}
		}
		if !orphaned {
			comments = append(comments, cg)
		}
	}
	f.Comments = comments

	return removedDecls
}

func writeFile(w io.Writer, fset *token.FileSet, f *ast.File) error {
	return format.Node(w, fset, f)
}

func isAPIServiceMethod(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return false
	}
	star, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	ident, ok := star.X.(*ast.Ident)
	return ok && ident.Name == "APIService"
}

// isNotImplementedStub reports whether fn's body references
// http.StatusNotImplemented, which every generated stub does and no real
// implementation should. Mirrors pkg/redfish/strip-redfish-routes's
// predicate of the same name.
func isNotImplementedStub(fn *ast.FuncDecl) bool {
	stub := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if stub {
			return false
		}
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "http" && sel.Sel.Name == "StatusNotImplemented" {
				stub = true
			}
		}
		return !stub
	})
	return stub
}
