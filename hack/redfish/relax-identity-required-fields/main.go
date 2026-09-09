// Command relax-identity-required-fields drops "Id" and "Name" from every
// generated model's UnmarshalJSON requiredProperties check.
//
// Every Redfish resource type embeds the spec's base Resource/Item schema,
// which marks Id and Name as required -- correct for a GET response, but
// openapi-generator's go-server template enforces the same requirement on
// every JSON body decoded into that type, including the ones real clients
// send to create or update a resource. Neither field is ever meant to come
// from the client: Id is either fixed by the URL path or assigned on
// creation, and Name is server-defined. A spec-compliant client that omits
// them (as real Redfish clients do, and as this project's own e2e tests
// already do for both ComputerSystem PATCH and Session creation) gets a
// 422 "field 'Id' is required" instead of the request it sent.
//
// Purpose-built request-body types (VirtualMediaV163InsertMediaRequestBody
// and friends) don't extend Resource and don't carry this problem -- their
// required fields are genuinely required. This only touches models whose
// UnmarshalJSON requires "Id" or "Name" in the first place.
//
// Usage (invoked from hack/redfish/generate.sh, after openapi-generator):
//
//	go run ./hack/redfish/relax-identity-required-fields \
//	    -dir pkg/generated/redfish/server
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
	"path/filepath"
	"strconv"
)

func main() {
	dir := flag.String("dir", "pkg/generated/redfish/server", "directory of generated models to fix up in place")
	flag.Parse()

	changed, err := run(*dir, "Id", "Name")
	if err != nil {
		fmt.Fprintf(os.Stderr, "relax-identity-required-fields: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "relax-identity-required-fields: relaxed %d model(s) in %s\n", changed, *dir)
}

func run(dir string, fields ...string) (int, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "model_*.go"))
	if err != nil {
		return 0, fmt.Errorf("glob %s: %w", dir, err)
	}

	changedCount := 0
	for _, path := range entries {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return changedCount, fmt.Errorf("parse %s: %w", path, err)
		}

		if !relaxRequiredProperties(f, fields...) {
			continue
		}
		changedCount++

		out, err := os.Create(path)
		if err != nil {
			return changedCount, fmt.Errorf("create %s: %w", path, err)
		}
		err = writeFile(out, fset, f)
		out.Close() //nolint:errcheck
		if err != nil {
			return changedCount, fmt.Errorf("write %s: %w", path, err)
		}
	}
	return changedCount, nil
}

// relaxRequiredProperties removes any of fields from every
// `requiredProperties := []string{...}` slice literal in f, and reports
// whether it changed anything.
func relaxRequiredProperties(f *ast.File, fields ...string) bool {
	drop := make(map[string]bool, len(fields))
	for _, field := range fields {
		drop[field] = true
	}

	changed := false
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}
		ident, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || ident.Name != "requiredProperties" {
			return true
		}
		lit, ok := assign.Rhs[0].(*ast.CompositeLit)
		if !ok {
			return true
		}

		kept := lit.Elts[:0]
		for _, elt := range lit.Elts {
			bl, ok := elt.(*ast.BasicLit)
			if !ok || bl.Kind != token.STRING {
				kept = append(kept, elt)
				continue
			}
			v, err := strconv.Unquote(bl.Value)
			if err != nil || !drop[v] {
				kept = append(kept, elt)
				continue
			}
			changed = true
		}
		lit.Elts = kept
		return true
	})
	return changed
}

func writeFile(w io.Writer, fset *token.FileSet, f *ast.File) error {
	return format.Node(w, fset, f)
}
