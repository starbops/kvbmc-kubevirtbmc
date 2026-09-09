package redfish

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
)

// sessionCreateRoute is the generated route name for
// POST /redfish/v1/SessionService/Sessions.
const sessionCreateRoute = "RedfishV1SessionServiceSessionsPost"

// sessionTokenFilter moves the session token out of the JSON response body
// and into the X-Auth-Token header, and sets Location to the new session's
// URI, per the Redfish spec (DSP0266): a successful session creation
// returns the session resource with no Token property, the client's
// session token in X-Auth-Token, and the resource's URI in Location.
//
// ImplResponse -- what every APIService method returns -- has no field for
// response headers, only a status code and a body, so this can't be done
// from api_service.go. The generated api_default.go wrapper used to be
// hand-patched to do this splice directly (protected via
// .openapi-generator-ignore); that stopped working once api_default.go had
// to regenerate along with the trimmed operation set. This wraps the route
// from the outside instead, so it's immune to regeneration: it only cares
// about the JSON the handler writes over the wire, not the generated
// model's Go shape.
type sessionTokenFilter struct {
	inner server.Router
}

func (f sessionTokenFilter) Routes() server.Routes {
	routes := f.inner.Routes()
	for name, route := range routes {
		if baseRouteName(name) == sessionCreateRoute {
			routes[name] = f.wrap(route)
		}
	}
	return routes
}

func (f sessionTokenFilter) OrderedRoutes() []server.Route {
	ordered := f.inner.OrderedRoutes()
	for i, route := range ordered {
		if baseRouteName(route.Name) == sessionCreateRoute {
			ordered[i] = f.wrap(route)
		}
	}
	return ordered
}

func (f sessionTokenFilter) wrap(route server.Route) server.Route {
	inner := route.HandlerFunc
	route.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		rec := &responseBuffer{header: make(http.Header), status: http.StatusOK}
		inner(rec, r)
		writeSessionResponse(w, rec)
	}
	return route
}

// writeSessionResponse copies rec's headers and status to w, extracting the
// session token into X-Auth-Token and setting Location if the buffered
// body has a "Token" property. Anything that doesn't decode as a JSON
// object with a Token property (error responses, in particular) is passed
// through unchanged.
func writeSessionResponse(w http.ResponseWriter, rec *responseBuffer) {
	for key, values := range rec.header {
		w.Header()[key] = values
	}

	var body map[string]any
	if err := json.Unmarshal(rec.body.Bytes(), &body); err != nil {
		w.WriteHeader(rec.status)
		_, _ = w.Write(rec.body.Bytes())
		return
	}

	token, ok := body["Token"].(string)
	if !ok || token == "" {
		w.WriteHeader(rec.status)
		_, _ = w.Write(rec.body.Bytes())
		return
	}
	delete(body, "Token")

	w.Header().Set("X-Auth-Token", token)
	if id, ok := body["Id"].(string); ok {
		w.Header().Set("Location", fmt.Sprintf("/redfish/v1/SessionService/Sessions/%s", id))
	}

	out, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(rec.status)
		_, _ = w.Write(rec.body.Bytes())
		return
	}
	w.WriteHeader(rec.status)
	_, _ = w.Write(out)
}

// responseBuffer is a minimal http.ResponseWriter that captures a handler's
// output instead of sending it, so it can be inspected and adjusted before
// anything reaches the real client.
type responseBuffer struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *responseBuffer) Header() http.Header         { return r.header }
func (r *responseBuffer) Write(b []byte) (int, error) { return r.body.Write(b) }
func (r *responseBuffer) WriteHeader(status int)      { r.status = status }
