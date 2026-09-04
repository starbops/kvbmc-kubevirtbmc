package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parseSource(t *testing.T, src string) (*ast.File, *token.FileSet) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "api_service.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return f, fset
}

func TestTrimStubs(t *testing.T) {
	src := `/*
 * Redfish
 *
 * This contains the definition of a Redfish service.
 */

package redfish

import (
	"errors"
	"net/http"

	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
)

type APIService struct{}

// RedfishV1Get -
func (s *APIService) RedfishV1Get(ctx context.Context) (server.ImplResponse, error) {
	return server.Response(200, nil), nil
}

// RedfishV1AccountServiceGet -
func (s *APIService) RedfishV1AccountServiceGet(ctx context.Context) (server.ImplResponse, error) {
	// TODO - update RedfishV1AccountServiceGet with the required logic for this service method.

	// TODO: Uncomment the next line to return response Response(200, {}) or use other options such as http.Ok ...
	// return Response(200, nil),nil

	return server.Response(http.StatusNotImplemented, nil), errors.New("RedfishV1AccountServiceGet method not implemented")
}
`
	f, fset := parseSource(t, src)

	removed := trimStubs(f)
	if removed != 1 {
		t.Fatalf("removed %d decls, want 1", removed)
	}

	var out strings.Builder
	if err := writeFile(&out, fset, f); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	got := out.String()

	if strings.Contains(got, "RedfishV1AccountServiceGet") {
		t.Errorf("stub method (or an orphaned comment from its body) survived:\n%s", got)
	}
	if strings.Contains(got, "TODO") {
		t.Errorf("a stub's internal TODO comment was orphaned instead of removed with it:\n%s", got)
	}
	if !strings.Contains(got, "func (s *APIService) RedfishV1Get(") {
		t.Errorf("real implementation was dropped:\n%s", got)
	}
	if !strings.Contains(got, "// RedfishV1Get -") {
		t.Errorf("real implementation's doc comment was dropped:\n%s", got)
	}
	if !strings.Contains(got, "This contains the definition of a Redfish service.") {
		t.Errorf("file header comment was dropped:\n%s", got)
	}
	if !strings.Contains(got, "type APIService struct") {
		t.Errorf("non-method declaration was dropped:\n%s", got)
	}
}

// TestRun_AgainstVendoredFile guards the real, currently-committed
// api_service.go: every method that survives trimming must be a real
// implementation, and the count must match the file's own stub/real split.
func TestRun_AgainstVendoredFile(t *testing.T) {
	const vendored = "../../../pkg/redfish/api_service.go"
	src, err := os.ReadFile(vendored)
	if err != nil {
		t.Skipf("vendored file not available: %v", err)
	}

	fset := token.NewFileSet()
	orig, err := parser.ParseFile(fset, vendored, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse vendored file: %v", err)
	}
	wantKept := 0
	for _, decl := range orig.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && isAPIServiceMethod(fn) && !isNotImplementedStub(fn) {
			wantKept++
		}
	}
	if wantKept == 0 {
		t.Fatal("found zero real implementations in the vendored file; is isNotImplementedStub still correct?")
	}

	scratch := filepath.Join(t.TempDir(), "api_service.go")
	if err := os.WriteFile(scratch, src, 0o644); err != nil {
		t.Fatalf("write scratch copy: %v", err)
	}
	if err := run(scratch); err != nil {
		t.Fatalf("run: %v", err)
	}

	out, err := os.ReadFile(scratch)
	if err != nil {
		t.Fatalf("read scratch copy: %v", err)
	}
	trimmed, err := parser.ParseFile(fset, scratch, out, 0)
	if err != nil {
		t.Fatalf("parse trimmed output: %v", err)
	}
	got := 0
	for _, decl := range trimmed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !isAPIServiceMethod(fn) {
			continue
		}
		got++
		if isNotImplementedStub(fn) {
			t.Errorf("stub method %s survived trimming", fn.Name.Name)
		}
	}
	if got != wantKept {
		t.Errorf("trimmed file has %d APIService methods, want %d (the real implementations)", got, wantKept)
	}
	// The original stubs run ~15-20 lines each including their TODO
	// comments; wantKept real methods plus boilerplate should be a small
	// fraction of the original file. This catches orphaned-comment bloat
	// (from a removed stub's internal TODOs surviving as floating
	// comments) that a pure decl/method count can't see. Real, kept
	// implementations may still contain their own leftover TODO text --
	// that's pre-existing untidiness in hand-written code, not something
	// this tool is responsible for cleaning up.
	if gotLines := strings.Count(string(out), "\n"); gotLines > 200*wantKept {
		t.Errorf("trimmed file has %d lines for %d real methods -- looks like leftover stub content", gotLines, wantKept)
	}
}
