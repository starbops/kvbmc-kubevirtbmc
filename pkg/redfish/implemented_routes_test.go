package redfish

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
)

func TestBaseRouteName(t *testing.T) {
	cases := map[string]string{
		"RedfishV1Get":                        "RedfishV1Get",
		"RedfishV1Get_0":                      "RedfishV1Get",
		"RedfishV1SystemsGet_12":              "RedfishV1SystemsGet",
		"RedfishV1SystemsComputerSystemId_":   "RedfishV1SystemsComputerSystemId_",
		"RedfishV1SystemsComputerSystemIdGet": "RedfishV1SystemsComputerSystemIdGet",
	}
	for in, want := range cases {
		if got := baseRouteName(in); got != want {
			t.Errorf("baseRouteName(%q) = %q, want %q", in, got, want)
		}
	}
}

// The generated set must cover the routes the BMC clients actually use, and
// must stay clear of stub-only operations.
func TestImplementedMethods(t *testing.T) {
	for _, name := range []string{
		"RedfishV1Get",
		"RedfishV1SessionServiceSessionsPost",
		"RedfishV1SystemsComputerSystemIdGet",
		"RedfishV1SystemsComputerSystemIdActionsComputerSystemResetPost",
	} {
		if !implementedMethods[name] {
			t.Errorf("implementedMethods missing %q; run 'make generate-implemented-routes'", name)
		}
	}
	if implementedMethods["RedfishV1MetadataGet"] {
		t.Errorf("RedfishV1MetadataGet is a 501 stub and must not be registered")
	}
}

// TestImplementedMethodsMatchAllowlist guards the other direction of the
// generation pipeline: every route pkg/redfish actually implements must be
// listed in hack/redfish/spec/implemented-operations.yaml, or
// trim-redfish-spec will strip it out of the spec and openapi-generator
// will stop generating it on the next `make generate-redfish-api`.
func TestImplementedMethodsMatchAllowlist(t *testing.T) {
	allowed, err := readAllowlist(t, "../../hack/redfish/spec/implemented-operations.yaml")
	if err != nil {
		t.Fatalf("read allowlist: %v", err)
	}

	controller := server.NewDefaultAPIController(NewAPIService("", "", nil))
	for routeName, route := range controller.Routes() {
		if !implementedMethods[baseRouteName(routeName)] {
			continue
		}
		op := route.Method + " " + route.Pattern
		if !allowed[op] {
			t.Errorf("%s (%s) is implemented but %q is missing from hack/redfish/spec/implemented-operations.yaml; "+
				"trim-redfish-spec will drop it and the next `make generate-redfish-api` will stop generating it", routeName, op, op)
		}
	}
}

// readAllowlist mirrors the "METHOD /path" parsing in
// hack/redfish/trim-redfish-spec, returning the set as "METHOD /path" keys.
func readAllowlist(t *testing.T, path string) (map[string]bool, error) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	allowed := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("malformed allowlist line: %q", line)
		}
		allowed[strings.ToUpper(fields[0])+" "+fields[1]] = true
	}
	return allowed, scanner.Err()
}
