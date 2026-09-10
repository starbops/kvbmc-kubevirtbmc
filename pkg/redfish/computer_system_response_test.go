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

func TestComputerSystemActionsFilter_AdvertisesResetAllowableValues(t *testing.T) {
	body := `{"Id":"1","Actions":{"#ComputerSystem.Reset":{"target":"/redfish/v1/Systems/1/Actions/ComputerSystem.Reset","title":"Reset"},"#ComputerSystem.SetDefaultBootOrder":{}}}`
	inner := fakeRouter{
		{Name: "RedfishV1SystemsComputerSystemIdGet", Method: "GET", Pattern: "/redfish/v1/Systems/{ComputerSystemId}",
			HandlerFunc: sessionCreateHandler(200, body)},
	}
	f := computerSystemActionsFilter{inner: inner}

	rec := httptest.NewRecorder()
	route := f.Routes()["RedfishV1SystemsComputerSystemIdGet"]
	route.HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1/Systems/1", nil))

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
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
	// The rest of the action object, and sibling actions, must survive.
	for _, want := range []string{
		`"target":"/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"`,
		`"title":"Reset"`,
		`"#ComputerSystem.SetDefaultBootOrder"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("response lost existing content %q: %s", want, got)
		}
	}
}

func TestComputerSystemActionsFilter_PassesThroughUnexpectedShapesUnchanged(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"no Actions key", `{"Id":"1"}`},
		{"Actions has no Reset action", `{"Id":"1","Actions":{"#ComputerSystem.Decommission":{}}}`},
		{"not JSON", `not json`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inner := fakeRouter{
				{Name: "RedfishV1SystemsComputerSystemIdGet", Method: "GET", Pattern: "/redfish/v1/Systems/{ComputerSystemId}",
					HandlerFunc: sessionCreateHandler(200, tc.body)},
			}
			f := computerSystemActionsFilter{inner: inner}

			rec := httptest.NewRecorder()
			route := f.Routes()["RedfishV1SystemsComputerSystemIdGet"]
			route.HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1/Systems/1", nil))

			if rec.Body.String() != tc.body {
				t.Errorf("body = %s, want unchanged: %s", rec.Body.String(), tc.body)
			}
		})
	}
}

func TestComputerSystemActionsFilter_OnlyWrapsComputerSystemGetRoute(t *testing.T) {
	untouchedBody := `{"Actions":{"#ComputerSystem.Reset":{}}}`
	inner := fakeRouter{
		{Name: "RedfishV1SystemsGet", Method: "GET", Pattern: "/redfish/v1/Systems",
			HandlerFunc: sessionCreateHandler(200, untouchedBody)},
	}
	f := computerSystemActionsFilter{inner: inner}

	rec := httptest.NewRecorder()
	route := f.Routes()["RedfishV1SystemsGet"]
	route.HandlerFunc.ServeHTTP(rec, httptest.NewRequest("GET", "/redfish/v1/Systems", nil))

	if rec.Body.String() != untouchedBody {
		t.Errorf("body = %s, want an unrelated route left completely untouched: %s", rec.Body.String(), untouchedBody)
	}
}

var _ server.Router = computerSystemActionsFilter{}

// TestComputerSystemGet_EndToEnd drives GET /redfish/v1/Systems/1 through
// the full router (routeFilter, computerSystemActionsFilter, authFilter,
// the real APIService), the same request the e2e regression test
// "should advertise supported reset action values" makes.
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
