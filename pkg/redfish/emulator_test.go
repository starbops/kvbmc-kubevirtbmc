package redfish

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
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
