package main

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppemployee"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppprodcategory"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppproduct"
)

// getRouter returns a mux providing the HTTP server's routes
func (srvApp *httpServerApplication) getRouter() http.Handler {

	// define env struct needed for route handlers
	apiEnv := lys.Env{
		ErrorLog:    srvApp.ErrorLog,
		Validate:    srvApp.Validate,
		GetOptions:  srvApp.GetOptions,
		PostOptions: srvApp.PostOptions,
	}

	// reduce default max # results from GET
	apiEnv.GetOptions.MaxPerPage = 100

	r := mux.NewRouter()

	// define middleware for all routes
	r.Use(secureHeaders)
	r.Use(srvApp.rejectBlockedIp)
	r.Use(srvApp.limit) // global rate limit based on IP

	r.NotFoundHandler = http.HandlerFunc(lys.NotFound())

	// no public login route: login and logout are faked by passing Employee-Email header

	// put all routes requiring auth behind "/a" for authed
	authedR := r.PathPrefix("/a").Subrouter()

	// define middleware for authed routes
	authedR.Use(srvApp.authenticate)
	authedR.Use(srvApp.logAuthedRequest) // must come after authentication

	// add subroutes into main router
	for _, subRoute := range srvApp.getSubRoutes(apiEnv) {
		subRouter := authedR.PathPrefix(subRoute.Url).Subrouter()
		_ = subRoute.RouteAdder(subRouter)
	}

	// apply CORS middleware to allow access to Vue app
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{srvApp.Config.UI.Url}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),

		// note use of Employee-Email for fake tenant auth
		handlers.AllowedHeaders([]string{"Accept", "Accept-Encoding", "Authorization", "Content-Length", "Content-Type", "Employee-Email", "X-CSRF-Token"}),

		handlers.ExposedHeaders([]string{"Content-Disposition"}),
		handlers.AllowCredentials(),
	)
	return (cors)(r)
}

// getSubRoutes returns all subroutes used by the server
func (srvApp *httpServerApplication) getSubRoutes(apiEnv lys.Env) []lys.SubRoute {

	return []lys.SubRoute{
		{Url: "/supplier", RouteAdder: srvApp.supplierRoutes(apiEnv)},
	}
}

func (srvApp *httpServerApplication) supplierRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {

		endpoint := "/employees"

		employeeStore := suppemployee.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, employeeStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, employeeStore)).Methods("GET")

		endpoint = "/product-categories"

		productCatStore := suppprodcategory.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, productCatStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, productCatStore)).Methods("GET")

		endpoint = "/products"

		productStore := suppproduct.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, productStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, productStore)).Methods("GET")
		r.HandleFunc(endpoint, lys.Post(apiEnv, productStore)).Methods("POST")
		r.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, productStore)).Methods("PUT")
		r.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, productStore)).Methods("PATCH")
		r.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, productStore)).Methods("DELETE")

		return r
	}
}
