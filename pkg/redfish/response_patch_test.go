package redfish

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func patchHandler(status int, body string, fn func(ctx context.Context)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if fn != nil {
			fn(r.Context())
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestPatchFilter_PatchResponseBody(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1SystemsComputerSystemIdGet", Method: "GET", Pattern: "/redfish/v1/Systems/{ComputerSystemId}",
			HandlerFunc: patchHandler(200, `{"Actions":{"#ComputerSystem.Reset":{"target":"x"}}}`, func(ctx context.Context) {
				PatchResponseBody(ctx, func(body map[string]any) {
					actions := body["Actions"].(map[string]any)
					reset := actions["#ComputerSystem.Reset"].(map[string]any)
					reset["ResetType@Redfish.AllowableValues"] = []string{"On", "ForceOff"}
				})
			})},
	}
	f := patchFilter{inner: inner}

	rec := httptest.NewRecorder()
	f.Routes()["RedfishV1SystemsComputerSystemIdGet"].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1/Systems/1", nil))

	got := rec.Body.String()
	for _, want := range []string{`"ResetType@Redfish.AllowableValues"`, `"On"`, `"ForceOff"`, `"target":"x"`} {
		if !strings.Contains(got, want) {
			t.Errorf("body missing %s: %s", want, got)
		}
	}
}

func TestPatchFilter_MultiplePatchesCompose(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: patchHandler(200, `{}`, func(ctx context.Context) {
			PatchResponseBody(ctx, func(body map[string]any) { body["First"] = 1 })
			PatchResponseBody(ctx, func(body map[string]any) { body["Second"] = 2 })
		})},
	}
	f := patchFilter{inner: inner}

	rec := httptest.NewRecorder()
	f.Routes()["RedfishV1Get"].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1", nil))

	got := rec.Body.String()
	if !strings.Contains(got, `"First":1`) || !strings.Contains(got, `"Second":2`) {
		t.Errorf("body = %s, want both patches applied", got)
	}
}

func TestPatchFilter_PassesThroughWhenNothingRequested(t *testing.T) {
	body := `{"Id":"1","Nested":{"A":true}}`
	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: patchHandler(201, body, nil)},
	}
	f := patchFilter{inner: inner}

	rec := httptest.NewRecorder()
	f.Routes()["RedfishV1Get"].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1", nil))

	if rec.Code != 201 {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	if rec.Body.String() != body {
		t.Errorf("body = %s, want byte-identical passthrough: %s", rec.Body.String(), body)
	}
	if len(rec.Header()) != 0 {
		t.Errorf("headers = %v, want none added", rec.Header())
	}
}

func TestPatchFilter_PatchOnNonJSONBodyPassesThroughUnchanged(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: patchHandler(500, "plain text error", func(ctx context.Context) {
			PatchResponseBody(ctx, func(body map[string]any) { body["Ignored"] = true })
		})},
	}
	f := patchFilter{inner: inner}

	rec := httptest.NewRecorder()
	f.Routes()["RedfishV1Get"].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1", nil))

	if rec.Body.String() != "plain text error" {
		t.Errorf("body = %s, want unchanged when it isn't JSON", rec.Body.String())
	}
}

func TestPatchFilter_OrderedRoutesAlsoWraps(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: patchHandler(200, `{}`, func(ctx context.Context) {
			PatchResponseBody(ctx, func(body map[string]any) { body["Ordered"] = true })
		})},
	}
	f := patchFilter{inner: inner}

	ordered := f.OrderedRoutes()
	if len(ordered) != 1 {
		t.Fatalf("OrderedRoutes() returned %d routes, want 1", len(ordered))
	}

	rec := httptest.NewRecorder()
	ordered[0].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1", nil))

	if !strings.Contains(rec.Body.String(), `"Ordered":true`) {
		t.Errorf("body = %s, want the patch applied via OrderedRoutes()", rec.Body.String())
	}
}

func TestPatchResponseBody_NoopWithoutPatchContext(t *testing.T) {
	// A caller using a plain context (e.g. a unit test that doesn't go
	// through patchFilter) must not panic -- it just has nothing to
	// collect into.
	PatchResponseBody(context.Background(), func(map[string]any) {})
}
