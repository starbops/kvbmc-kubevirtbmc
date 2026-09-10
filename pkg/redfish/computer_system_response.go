package redfish

import (
	"encoding/json"
	"net/http"

	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
)

// computerSystemGetRoute is the generated route name for
// GET /redfish/v1/Systems/{ComputerSystemId}.
const computerSystemGetRoute = "RedfishV1SystemsComputerSystemIdGet"

// resetAllowableValues are the ComputerSystem.Reset parameter values
// KubeVirtBMC accepts, advertised on every ComputerSystem GET via the
// "ResetType@Redfish.AllowableValues" OData annotation on the
// #ComputerSystem.Reset action -- see kubevirtbmc#204: clients need this to
// distinguish graceful from force operations before choosing one.
var resetAllowableValues = []server.ResourceResetType{
	server.RESOURCERESETTYPE_ON,
	server.RESOURCERESETTYPE_FORCE_OFF,
	server.RESOURCERESETTYPE_GRACEFUL_SHUTDOWN,
	server.RESOURCERESETTYPE_GRACEFUL_RESTART,
	server.RESOURCERESETTYPE_FORCE_RESTART,
}

// computerSystemActionsFilter advertises resetAllowableValues on every
// ComputerSystem GET response.
//
// openapi-generator's go-server template used to generate a
// ResetTypeRedfishAllowableValues field on the Reset action struct for
// this annotation. hack/redfish/generate.sh fetches ComputerSystem.v1_22_0
// live from redfish.dmtf.org on every run, and DMTF's currently served
// copy no longer round-trips that annotation into the generated model the
// same way -- whether the field exists again depends on what DMTF happens
// to be serving on a given day, which is exactly what silently broke this
// the first time. Advertising it here instead, from KubeVirtBMC's own
// fixed list of supported reset types, doesn't depend on the generated
// model having anywhere to put it at all: this only cares about the JSON a
// handler writes over the wire, the same fix already applied to session
// creation's headers.
type computerSystemActionsFilter struct {
	inner server.Router
}

func (f computerSystemActionsFilter) Routes() server.Routes {
	routes := f.inner.Routes()
	for name, route := range routes {
		if baseRouteName(name) == computerSystemGetRoute {
			routes[name] = f.wrap(route)
		}
	}
	return routes
}

func (f computerSystemActionsFilter) OrderedRoutes() []server.Route {
	ordered := f.inner.OrderedRoutes()
	for i, route := range ordered {
		if baseRouteName(route.Name) == computerSystemGetRoute {
			ordered[i] = f.wrap(route)
		}
	}
	return ordered
}

func (f computerSystemActionsFilter) wrap(route server.Route) server.Route {
	inner := route.HandlerFunc
	route.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		rec := &responseBuffer{header: make(http.Header), status: http.StatusOK}
		inner(rec, r)
		writeComputerSystemResponse(w, rec)
	}
	return route
}

// writeComputerSystemResponse copies rec's headers and status to w, adding
// the ResetType@Redfish.AllowableValues annotation to the
// #ComputerSystem.Reset action if the buffered body has one. Anything that
// doesn't decode into that exact shape (error responses, in particular) is
// passed through unchanged.
func writeComputerSystemResponse(w http.ResponseWriter, rec *responseBuffer) {
	for key, values := range rec.header {
		w.Header()[key] = values
	}

	if out, ok := withResetAllowableValues(rec.body.Bytes()); ok {
		w.WriteHeader(rec.status)
		_, _ = w.Write(out)
		return
	}
	w.WriteHeader(rec.status)
	_, _ = w.Write(rec.body.Bytes())
}

func withResetAllowableValues(body []byte) ([]byte, bool) {
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, false
	}
	actions, ok := decoded["Actions"].(map[string]any)
	if !ok {
		return nil, false
	}
	reset, ok := actions["#ComputerSystem.Reset"].(map[string]any)
	if !ok {
		return nil, false
	}

	reset["ResetType@Redfish.AllowableValues"] = resetAllowableValues
	out, err := json.Marshal(decoded)
	if err != nil {
		return nil, false
	}
	return out, true
}
