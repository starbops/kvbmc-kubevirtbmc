package redfish

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
)

// Some Redfish responses need a body property a generated model doesn't
// happen to have (an OData "@Redfish.AllowableValues" annotation -- whether
// it does depends on what redfish.dmtf.org happens to be serving the day
// generate.sh last ran, which has already broken twice). PatchResponseBody
// lets an APIService method ask for one, using the Go values it already
// has on hand, without knowing or caring how the response actually gets
// assembled. (Response headers don't need this: openapi-generator's
// addResponseHeaders option gives ImplResponse a Headers field directly --
// see server.ResponseWithHeaders.)
//
// patchFilter is the other half: applied once, to every route, it buffers
// the handler's response, applies whatever patches were requested via ctx,
// and copies both the (possibly patched) body and whatever headers the
// handler set -- via ImplResponse.Headers or otherwise -- onto the real
// response. A handler that calls neither PatchResponseBody nor
// ResponseWithHeaders costs it nothing beyond a small buffer-and-copy.
type patchContextKey struct{}

type patchState struct {
	patches []func(body map[string]any)
}

// PatchResponseBody registers a function to mutate the response body,
// decoded as a generic JSON object, before it's written to the client.
// Multiple patches on the same request all run, in registration order. A
// no-op if ctx didn't come from a request routed through patchFilter (e.g.
// a unit test using context.Background()); also silently skipped if the
// body isn't a JSON object (an error response, for instance), since
// there's nothing to patch.
func PatchResponseBody(ctx context.Context, patch func(body map[string]any)) {
	if s := patchesFrom(ctx); s != nil {
		s.patches = append(s.patches, patch)
	}
}

func patchesFrom(ctx context.Context) *patchState {
	s, _ := ctx.Value(patchContextKey{}).(*patchState)
	return s
}

// patchFilter applies response body patches requested via
// PatchResponseBody. See the package doc comment above for why this
// exists.
type patchFilter struct {
	inner server.Router
}

func (f patchFilter) Routes() server.Routes {
	routes := f.inner.Routes()
	wrapped := make(server.Routes, len(routes))
	for name, route := range routes {
		wrapped[name] = f.wrap(route)
	}
	return wrapped
}

func (f patchFilter) OrderedRoutes() []server.Route {
	ordered := f.inner.OrderedRoutes()
	wrapped := make([]server.Route, len(ordered))
	for i, route := range ordered {
		wrapped[i] = f.wrap(route)
	}
	return wrapped
}

func (f patchFilter) wrap(route server.Route) server.Route {
	inner := route.HandlerFunc
	route.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		state := &patchState{}
		r = r.WithContext(context.WithValue(r.Context(), patchContextKey{}, state))

		rec := &responseBuffer{header: make(http.Header), status: http.StatusOK}
		inner(rec, r)

		for key, values := range rec.header {
			w.Header()[key] = values
		}

		body := rec.body.Bytes()
		if len(state.patches) > 0 {
			if patched, ok := applyPatches(body, state.patches); ok {
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
