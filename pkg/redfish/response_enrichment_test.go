package redfish

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func enrichmentHandler(status int, body string, fn func(ctx context.Context)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if fn != nil {
			fn(r.Context())
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestEnrichmentFilter_AddResponseHeader(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: enrichmentHandler(200, `{"Id":"1"}`, func(ctx context.Context) {
			AddResponseHeader(ctx, "X-Auth-Token", "secret-token")
			AddResponseHeader(ctx, "Location", "/redfish/v1/SessionService/Sessions/1")
		})},
	}
	f := enrichmentFilter{inner: inner}

	rec := httptest.NewRecorder()
	f.Routes()["RedfishV1Get"].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1", nil))

	if got := rec.Header().Get("X-Auth-Token"); got != "secret-token" {
		t.Errorf("X-Auth-Token = %q, want %q", got, "secret-token")
	}
	if got := rec.Header().Get("Location"); got != "/redfish/v1/SessionService/Sessions/1" {
		t.Errorf("Location = %q, want %q", got, "/redfish/v1/SessionService/Sessions/1")
	}
	if rec.Body.String() != `{"Id":"1"}` {
		t.Errorf("body = %s, want it unchanged", rec.Body.String())
	}
}

func TestEnrichmentFilter_PatchResponseBody(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1SystemsComputerSystemIdGet", Method: "GET", Pattern: "/redfish/v1/Systems/{ComputerSystemId}",
			HandlerFunc: enrichmentHandler(200, `{"Actions":{"#ComputerSystem.Reset":{"target":"x"}}}`, func(ctx context.Context) {
				PatchResponseBody(ctx, func(body map[string]any) {
					actions := body["Actions"].(map[string]any)
					reset := actions["#ComputerSystem.Reset"].(map[string]any)
					reset["ResetType@Redfish.AllowableValues"] = []string{"On", "ForceOff"}
				})
			})},
	}
	f := enrichmentFilter{inner: inner}

	rec := httptest.NewRecorder()
	f.Routes()["RedfishV1SystemsComputerSystemIdGet"].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1/Systems/1", nil))

	got := rec.Body.String()
	for _, want := range []string{`"ResetType@Redfish.AllowableValues"`, `"On"`, `"ForceOff"`, `"target":"x"`} {
		if !strings.Contains(got, want) {
			t.Errorf("body missing %s: %s", want, got)
		}
	}
}

func TestEnrichmentFilter_MultiplePatchesCompose(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: enrichmentHandler(200, `{}`, func(ctx context.Context) {
			PatchResponseBody(ctx, func(body map[string]any) { body["First"] = 1 })
			PatchResponseBody(ctx, func(body map[string]any) { body["Second"] = 2 })
		})},
	}
	f := enrichmentFilter{inner: inner}

	rec := httptest.NewRecorder()
	f.Routes()["RedfishV1Get"].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1", nil))

	got := rec.Body.String()
	if !strings.Contains(got, `"First":1`) || !strings.Contains(got, `"Second":2`) {
		t.Errorf("body = %s, want both patches applied", got)
	}
}

func TestEnrichmentFilter_PassesThroughWhenNothingRequested(t *testing.T) {
	body := `{"Id":"1","Nested":{"A":true}}`
	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: enrichmentHandler(201, body, nil)},
	}
	f := enrichmentFilter{inner: inner}

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

func TestEnrichmentFilter_PatchOnNonJSONBodyPassesThroughUnchanged(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: enrichmentHandler(500, "plain text error", func(ctx context.Context) {
			PatchResponseBody(ctx, func(body map[string]any) { body["Ignored"] = true })
		})},
	}
	f := enrichmentFilter{inner: inner}

	rec := httptest.NewRecorder()
	f.Routes()["RedfishV1Get"].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1", nil))

	if rec.Body.String() != "plain text error" {
		t.Errorf("body = %s, want unchanged when it isn't JSON", rec.Body.String())
	}
}

func TestEnrichmentFilter_OrderedRoutesAlsoWraps(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: enrichmentHandler(200, `{}`, func(ctx context.Context) {
			AddResponseHeader(ctx, "X-Test", "ordered")
		})},
	}
	f := enrichmentFilter{inner: inner}

	ordered := f.OrderedRoutes()
	if len(ordered) != 1 {
		t.Fatalf("OrderedRoutes() returned %d routes, want 1", len(ordered))
	}

	rec := httptest.NewRecorder()
	ordered[0].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1", nil))

	if got := rec.Header().Get("X-Test"); got != "ordered" {
		t.Errorf("X-Test = %q, want %q", got, "ordered")
	}
}

func TestAddResponseHeader_NoopWithoutEnrichmentContext(t *testing.T) {
	// A caller using a plain context (e.g. a unit test that doesn't go
	// through enrichmentFilter) must not panic -- it just has nothing to
	// collect into.
	AddResponseHeader(context.Background(), "X-Test", "value")
	PatchResponseBody(context.Background(), func(map[string]any) {})
}
