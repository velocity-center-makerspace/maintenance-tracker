package router

import (
	"database/sql"
	"net/http"

	"github.com/velocity-center-makerspace/maintenance-tracker/db"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/config"
)

type Router interface {
	AddHandlers()
	AddRoute(pattern string, handler DepsHandlerFunc)
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type DepsHandlerFunc func(
	deps Dependencies, w http.ResponseWriter, req *http.Request)

type Dependencies struct {
	DB  *sql.DB
	Qry *db.Queries
	Env *config.Environment
}

type route struct {
	pattern string
	handler DepsHandlerFunc
}

type router struct {
	deps   Dependencies
	mux    *http.ServeMux
	routes []route
}

func New(deps Dependencies) Router {
	return &router{
		deps: deps,
		mux:  http.NewServeMux(),
	}
}

func (r *router) AddHandlers() {
	for _, rt := range r.routes {
		h := rt.handler
		r.mux.HandleFunc(
			rt.pattern,
			func(w http.ResponseWriter, req *http.Request) {
				h(r.deps, w, req)
			})
	}
}

func (r *router) AddRoute(pattern string, handler DepsHandlerFunc) {
	r.routes = append(
		r.routes,
		route{
			pattern: pattern,
			handler: handler,
		},
	)
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
