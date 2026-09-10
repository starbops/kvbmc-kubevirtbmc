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

// fakeRouter is a minimal server.Router test double built from a fixed set
// of routes, in the order given.
type fakeRouter []server.Route

func (f fakeRouter) Routes() server.Routes {
	routes := make(server.Routes, len(f))
	for _, r := range f {
		routes[r.Name] = r
	}
	return routes
}

func (f fakeRouter) OrderedRoutes() []server.Route {
	return append([]server.Route(nil), f...)
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestRouteFilter(t *testing.T) {
	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: okHandler},
		{Name: "RedfishV1AccountServiceGet", Method: "GET", Pattern: "/redfish/v1/AccountService", HandlerFunc: okHandler},
	}
	f := routeFilter{inner}

	routes := f.Routes()
	if _, ok := routes["RedfishV1Get"]; !ok {
		t.Error("Routes() dropped an implemented route")
	}
	if _, ok := routes["RedfishV1AccountServiceGet"]; ok {
		t.Error("Routes() kept a route with no real implementation")
	}

	ordered := f.OrderedRoutes()
	if len(ordered) != 1 || ordered[0].Name != "RedfishV1Get" {
		t.Errorf("OrderedRoutes() = %v, want exactly the implemented route", ordered)
	}
}

func TestAuthFilter(t *testing.T) {
	middlewareCalls := 0
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalls++
			next.ServeHTTP(w, r)
		})
	}

	inner := fakeRouter{
		{Name: "RedfishV1Get", Method: "GET", Pattern: "/redfish/v1", HandlerFunc: okHandler},
		{Name: "RedfishV1Get_0", Method: "GET", Pattern: "/redfish/v1/", HandlerFunc: okHandler},
		{Name: "RedfishV1SessionServiceSessionsPost", Method: "POST", Pattern: "/redfish/v1/SessionService/Sessions", HandlerFunc: okHandler},
		{Name: "RedfishV1SystemsGet", Method: "GET", Pattern: "/redfish/v1/Systems", HandlerFunc: okHandler},
	}
	f := authFilter{inner: inner, middleware: middleware}

	cases := []struct {
		name        string
		wantWrapped bool
		fromOrdered bool
	}{
		{"RedfishV1Get", false, false},
		{"RedfishV1Get_0", false, true},
		{"RedfishV1SessionServiceSessionsPost", false, false},
		{"RedfishV1SystemsGet", true, false},
	}

	routes := f.Routes()
	ordered := make(map[string]server.Route, len(f.OrderedRoutes()))
	for _, r := range f.OrderedRoutes() {
		ordered[r.Name] = r
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			route, ok := routes[tc.name]
			if !ok {
				t.Fatalf("Routes() missing %q", tc.name)
			}
			if orderedRoute, ok := ordered[tc.name]; !ok {
				t.Fatalf("OrderedRoutes() missing %q", tc.name)
			} else if tc.fromOrdered {
				route = orderedRoute
			}

			before := middlewareCalls
			rec := httptest.NewRecorder()
			route.HandlerFunc.ServeHTTP(rec, httptest.NewRequest(route.Method, route.Pattern, nil))

			called := middlewareCalls > before
			if called != tc.wantWrapped {
				t.Errorf("middleware invoked = %v, want %v", called, tc.wantWrapped)
			}
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d (handler must still run)", rec.Code, http.StatusOK)
			}
		})
	}
}

// TestSessionCreate_EndToEnd drives the exact request from
// kubevirtbmc#297's regression report through the real router
// (routeFilter, enrichmentFilter, authFilter, and the real
// APIService.RedfishV1SessionServiceSessionsPost), the same way the curl
// reproduction did:
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

// TestComputerSystemGet_EndToEnd drives GET /redfish/v1/Systems/1 through
// the full router, the same request the e2e regression test "should
// advertise supported reset action values" (kubevirtbmc#204) makes.
func TestComputerSystemGet_EndToEnd(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	mockRM := resourcemanager.NewMockResourceManager(ctl)
	mockRM.EXPECT().
		GetComputerSystem(gomock.Any()).
		Return(resourcemanager.NewComputerSystem("1", "default/testvm", server.RESOURCEPOWERSTATE_ON), nil)
	mockRM.EXPECT().GetBootFlags(gomock.Any()).Return(nil, nil)

	router := newRouter(testUsername, testPassword, mockRM)

	req := httptest.NewRequest("GET", "/redfish/v1/Systems/1", nil)
	req.SetBasicAuth(testUsername, testPassword)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	got := rec.Body.String()
	if !strings.Contains(got, `"ResetType@Redfish.AllowableValues"`) {
		t.Fatalf("missing ResetType@Redfish.AllowableValues annotation: %s", got)
	}
	for _, want := range []string{"On", "ForceOff", "GracefulShutdown", "GracefulRestart", "ForceRestart"} {
		if !strings.Contains(got, `"`+want+`"`) {
			t.Errorf("missing allowable value %q: %s", want, got)
		}
	}
}
