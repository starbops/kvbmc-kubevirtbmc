package redfish

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
	"kubevirt.io/kubevirtbmc/pkg/resourcemanager"
	"kubevirt.io/kubevirtbmc/pkg/session"
)

type Emulator struct {
	ctx  context.Context
	port int

	bmcUser     string
	bmcPassword string

	wg     sync.WaitGroup
	server *http.Server
}

// newRouter builds the full Redfish route table: implemented routes only,
// response enrichment applied for whatever a handler asked for via
// AddResponseHeader/PatchResponseBody, and authentication required on
// everything but the public routes.
func newRouter(bmcUser, bmcPassword string, resourceManager resourcemanager.ResourceManager) http.Handler {
	apiService := NewAPIService(bmcUser, bmcPassword, resourceManager)
	apiController := server.NewDefaultAPIController(apiService, server.WithDefaultAPIErrorHandler(recordingErrorHandler))
	return server.NewRouter(authFilter{
		inner: enrichmentFilter{
			inner: routeFilter{apiController},
		},
		middleware: session.AuthMiddleware(bmcUser, bmcPassword),
	})
}

func NewEmulator(ctx context.Context, port int, bmcUser string, bmcPassword string, resourceManager resourcemanager.ResourceManager) *Emulator {
	router := newRouter(bmcUser, bmcPassword, resourceManager)

	// Mount /healthz outside the access-log wrapper so readiness probes stay silent.
	root := http.NewServeMux()
	root.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	root.Handle("/", accessLog(router))

	return &Emulator{
		ctx:         ctx,
		port:        port,
		bmcUser:     bmcUser,
		bmcPassword: bmcPassword,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: root,
		},
	}
}

func (e *Emulator) Run() error {
	e.wg.Add(1)

	go func() {
		defer e.wg.Done()

		if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println(err)
		}
	}()

	return nil
}

func (e *Emulator) Stop() {
	if err := e.server.Shutdown(e.ctx); err != nil {
		fmt.Println(err)
	}
	e.wg.Wait()
	logrus.Info("Redfish emulator gracefully stopped")
}

//go:generate go run ./strip-redfish-routes -input api_service.go -output implemented_routes_gen.go

// routeFilter registers only the routes backed by a real implementation. The
// generated route table holds one entry per Redfish operation (~4000, all but
// a handful answering 501), and gorilla/mux compiles every registered pattern
// into a regexp that stays live for the process lifetime — ~50MB of heap per
// agent pod, see kubevirtbmc/kubevirtbmc#264.
type routeFilter struct {
	inner server.Router
}

func (f routeFilter) Routes() server.Routes {
	routes := f.inner.Routes()
	for name := range routes {
		if !implementedMethods[baseRouteName(name)] {
			delete(routes, name)
		}
	}
	return routes
}

func (f routeFilter) OrderedRoutes() []server.Route {
	ordered := f.inner.OrderedRoutes()
	kept := ordered[:0]
	for _, route := range ordered {
		if implementedMethods[baseRouteName(route.Name)] {
			kept = append(kept, route)
		}
	}
	return kept
}

// publicRoutes are the routes a Redfish client must be able to reach before
// it has a session: discovering the service root, and creating the session
// itself. Everything else requires a valid session or basic-auth
// credentials. This mirrors the auth split the OpenAPI generator's go-server
// template used to hard-code directly into routers.go; that file is now
// fully generated (and regenerated), so the split lives here instead, where
// a generator upgrade can't silently drop it.
var publicRoutes = map[string]bool{
	"RedfishV1Get":                        true,
	"RedfishV1SessionServiceSessionsPost": true,
}

// authFilter requires a valid session for every route except publicRoutes.
type authFilter struct {
	inner      server.Router
	middleware func(http.Handler) http.Handler
}

func (f authFilter) Routes() server.Routes {
	routes := f.inner.Routes()
	wrapped := make(server.Routes, len(routes))
	for name, route := range routes {
		wrapped[name] = f.wrap(route)
	}
	return wrapped
}

func (f authFilter) OrderedRoutes() []server.Route {
	ordered := f.inner.OrderedRoutes()
	wrapped := make([]server.Route, len(ordered))
	for i, route := range ordered {
		wrapped[i] = f.wrap(route)
	}
	return wrapped
}

func (f authFilter) wrap(route server.Route) server.Route {
	if publicRoutes[baseRouteName(route.Name)] {
		return route
	}
	route.HandlerFunc = f.middleware(route.HandlerFunc).ServeHTTP
	return route
}

// baseRouteName strips the _N suffix the OpenAPI generator appends to alias
// paths of the same operation (e.g. RedfishV1Get_0 -> RedfishV1Get), so alias
// routes inherit the implementation status of their canonical route.
func baseRouteName(name string) string {
	i := strings.LastIndexByte(name, '_')
	if i < 0 || i+1 == len(name) {
		return name
	}
	for _, c := range name[i+1:] {
		if c < '0' || c > '9' {
			return name
		}
	}
	return name[:i]
}
