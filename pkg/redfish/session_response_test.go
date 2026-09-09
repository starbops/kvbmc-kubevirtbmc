package redfish

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
	"kubevirt.io/kubevirtbmc/pkg/resourcemanager"
)

func sessionCreateHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestSessionTokenFilter_ExtractsTokenAndSetsHeaders(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1SessionServiceSessionsPost", Method: "POST", Pattern: "/redfish/v1/SessionService/Sessions",
			HandlerFunc: sessionCreateHandler(201, `{"Id":"abc123","Name":"User Session","Token":"secret-token","UserName":"admin"}`)},
	}
	f := sessionTokenFilter{inner: inner}

	rec := httptest.NewRecorder()
	route := f.Routes()["RedfishV1SessionServiceSessionsPost"]
	route.HandlerFunc.ServeHTTP(rec, httptest.NewRequest("POST", "/redfish/v1/SessionService/Sessions", nil))

	if got := rec.Header().Get("X-Auth-Token"); got != "secret-token" {
		t.Errorf("X-Auth-Token = %q, want %q", got, "secret-token")
	}
	if got := rec.Header().Get("Location"); got != "/redfish/v1/SessionService/Sessions/abc123" {
		t.Errorf("Location = %q, want %q", got, "/redfish/v1/SessionService/Sessions/abc123")
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=UTF-8" {
		t.Errorf("Content-Type = %q, want it preserved from the inner handler", got)
	}
	if rec.Code != 201 {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "Token") {
		t.Errorf("response body still contains Token: %s", body)
	}
	for _, want := range []string{`"Id":"abc123"`, `"Name":"User Session"`, `"UserName":"admin"`} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %s: %s", want, body)
		}
	}
}

func TestSessionTokenFilter_PassesThroughErrorResponsesUnchanged(t *testing.T) {
	errorBody := `{"error":{"code":"Base.1.0.GeneralError","message":"invalid username or password"}}`
	inner := fakeRouter{
		{Name: "RedfishV1SessionServiceSessionsPost", Method: "POST", Pattern: "/redfish/v1/SessionService/Sessions",
			HandlerFunc: sessionCreateHandler(401, errorBody)},
	}
	f := sessionTokenFilter{inner: inner}

	rec := httptest.NewRecorder()
	route := f.Routes()["RedfishV1SessionServiceSessionsPost"]
	route.HandlerFunc.ServeHTTP(rec, httptest.NewRequest("POST", "/redfish/v1/SessionService/Sessions", nil))

	if rec.Code != 401 {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	if rec.Body.String() != errorBody {
		t.Errorf("body = %s, want it passed through byte-for-byte: %s", rec.Body.String(), errorBody)
	}
	if got := rec.Header().Get("X-Auth-Token"); got != "" {
		t.Errorf("X-Auth-Token should not be set on an error response, got %q", got)
	}
	if got := rec.Header().Get("Location"); got != "" {
		t.Errorf("Location should not be set on an error response, got %q", got)
	}
}

func TestSessionTokenFilter_OnlyWrapsSessionCreateRoute(t *testing.T) {
	untouchedBody := `{"Token":"should-not-be-touched"}`
	inner := fakeRouter{
		{Name: "RedfishV1SystemsGet", Method: "GET", Pattern: "/redfish/v1/Systems",
			HandlerFunc: sessionCreateHandler(200, untouchedBody)},
	}
	f := sessionTokenFilter{inner: inner}

	rec := httptest.NewRecorder()
	route := f.Routes()["RedfishV1SystemsGet"]
	route.HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1/Systems", nil))

	if rec.Body.String() != untouchedBody {
		t.Errorf("body = %s, want an unrelated route left completely untouched: %s", rec.Body.String(), untouchedBody)
	}
	if got := rec.Header().Get("X-Auth-Token"); got != "" {
		t.Errorf("X-Auth-Token should not be set for an unrelated route, got %q", got)
	}
}

func TestSessionTokenFilter_OrderedRoutesAlsoWraps(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1SessionServiceSessionsPost", Method: "POST", Pattern: "/redfish/v1/SessionService/Sessions",
			HandlerFunc: sessionCreateHandler(201, `{"Id":"xyz","Token":"tok"}`)},
	}
	f := sessionTokenFilter{inner: inner}

	ordered := f.OrderedRoutes()
	if len(ordered) != 1 {
		t.Fatalf("OrderedRoutes() returned %d routes, want 1", len(ordered))
	}

	rec := httptest.NewRecorder()
	ordered[0].HandlerFunc.ServeHTTP(rec, httptest.NewRequest("POST", "/redfish/v1/SessionService/Sessions", nil))

	if got := rec.Header().Get("X-Auth-Token"); got != "tok" {
		t.Errorf("X-Auth-Token = %q, want %q", got, "tok")
	}
}

var _ server.Router = sessionTokenFilter{}

// TestSessionCreate_EndToEnd drives the exact request from the reported
// bug through the real router (routeFilter, sessionTokenFilter, authFilter,
// and the real APIService.RedfishV1SessionServiceSessionsPost), the same
// way the curl reproduction did:
//
//	curl -u admin:supersecret -X POST .../redfish/v1/SessionService/Sessions \
//	    -d '{"UserName": "admin", "Password": "supersecret"}'
func TestSessionCreate_EndToEnd(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	mockRM := resourcemanager.NewMockResourceManager(ctl)

	router := newRouter(testUsername, testPassword, mockRM)

	req := httptest.NewRequest(
		"POST", "/redfish/v1/SessionService/Sessions",
		strings.NewReader(`{"UserName": "admin", "Password": "admin123"}`),
	)
	req.SetBasicAuth(testUsername, testPassword)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if got := rec.Header().Get("X-Auth-Token"); got == "" {
		t.Error("X-Auth-Token header is missing")
	}
	if got := rec.Header().Get("Location"); got == "" {
		t.Error("Location header is missing")
	}
	if strings.Contains(rec.Body.String(), "Token") {
		t.Errorf("response body still contains the session token: %s", rec.Body.String())
	}
}
