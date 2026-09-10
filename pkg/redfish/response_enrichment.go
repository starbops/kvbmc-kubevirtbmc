package redfish

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
)

// Some Redfish responses need something ImplResponse has no field for: a
// header (a session's token, its Location), or a body property a generated
// model doesn't happen to have (an OData "@Redfish.AllowableValues"
// annotation -- whether it does depends on what redfish.dmtf.org happens
// to be serving the day generate.sh last ran, which has already broken
// twice). AddResponseHeader and PatchResponseBody let an APIService method
// ask for either, using the Go values it already has on hand, without
// knowing or caring how the response actually gets assembled.
//
// enrichmentFilter is the other half: applied once, to every route, it
// collects whatever a handler asked for via ctx and applies it to the real
// response afterward. A handler that calls neither helper costs it nothing
// beyond a small buffer-and-copy. Adding a new case where a generated
// model can't express something Redfish requires means calling one of
// these two functions at the point that already has the data -- never a
// new wrapper type, a new route-name constant, or another level of
// nesting in newRouter.
type responseEnrichmentKey struct{}

type responseEnrichment struct {
	headers http.Header
	patches []func(body map[string]any)
}

// AddResponseHeader adds a header to the eventual HTTP response for the
// request behind ctx. A no-op if ctx didn't come from a request routed
// through enrichmentFilter (e.g. a unit test using context.Background()).
func AddResponseHeader(ctx context.Context, key, value string) {
	if e := enrichmentFrom(ctx); e != nil {
		e.headers.Add(key, value)
	}
}

// PatchResponseBody registers a function to mutate the response body,
// decoded as a generic JSON object, before it's written to the client.
// Multiple patches on the same request all run, in registration order. A
// no-op under the same conditions as AddResponseHeader; also silently
// skipped if the body isn't a JSON object (an error response, for
// instance), since there's nothing to patch.
func PatchResponseBody(ctx context.Context, patch func(body map[string]any)) {
	if e := enrichmentFrom(ctx); e != nil {
		e.patches = append(e.patches, patch)
	}
}

func enrichmentFrom(ctx context.Context) *responseEnrichment {
	e, _ := ctx.Value(responseEnrichmentKey{}).(*responseEnrichment)
	return e
}

// enrichmentFilter applies response enrichment requested via
// AddResponseHeader/PatchResponseBody. See the package doc comment above
// for why this exists.
type enrichmentFilter struct {
	inner server.Router
}

func (f enrichmentFilter) Routes() server.Routes {
	routes := f.inner.Routes()
	wrapped := make(server.Routes, len(routes))
	for name, route := range routes {
		wrapped[name] = f.wrap(route)
	}
	return wrapped
}

func (f enrichmentFilter) OrderedRoutes() []server.Route {
	ordered := f.inner.OrderedRoutes()
	wrapped := make([]server.Route, len(ordered))
	for i, route := range ordered {
		wrapped[i] = f.wrap(route)
	}
	return wrapped
}

func (f enrichmentFilter) wrap(route server.Route) server.Route {
	inner := route.HandlerFunc
	route.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		enrichment := &responseEnrichment{headers: make(http.Header)}
		r = r.WithContext(context.WithValue(r.Context(), responseEnrichmentKey{}, enrichment))

		rec := &responseBuffer{header: make(http.Header), status: http.StatusOK}
		inner(rec, r)

		for key, values := range rec.header {
			w.Header()[key] = values
		}
		for key, values := range enrichment.headers {
			w.Header()[key] = values
		}

		body := rec.body.Bytes()
		if len(enrichment.patches) > 0 {
			if patched, ok := applyPatches(body, enrichment.patches); ok {
				body = patched
			}
		}
		w.WriteHeader(rec.status)
		_, _ = w.Write(body)
	}
	return route
}

// applyPatches decodes body as a JSON object, runs every patch against it
// in order, and re-encodes it. Reports false (leaving body untouched) if
// it doesn't decode as a JSON object in the first place.
func applyPatches(body []byte, patches []func(map[string]any)) ([]byte, bool) {
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, false
	}
	for _, patch := range patches {
		patch(decoded)
	}
	out, err := json.Marshal(decoded)
	if err != nil {
		return nil, false
	}
	return out, true
}

// responseBuffer is a minimal http.ResponseWriter that captures a
// handler's output instead of sending it, so it can be inspected and
// adjusted before anything reaches the real client.
type responseBuffer struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *responseBuffer) Header() http.Header         { return r.header }
func (r *responseBuffer) Write(b []byte) (int, error) { return r.body.Write(b) }
func (r *responseBuffer) WriteHeader(status int)      { r.status = status }
