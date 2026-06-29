package main

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/connectors/aws/stores/awsapicall"
	"github.com/loveyourstack/connectors/ecb/stores/ecbapicall"
	"github.com/loveyourstack/connectors/ecb/stores/ecbexchangerate"
	"github.com/loveyourstack/connectors/maxmind/stores/mmapicall"
	"github.com/loveyourstack/connectors/maxmind/stores/mmlocation"
	"github.com/loveyourstack/connectors/maxmind/stores/mmnetwork"
	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys-ref/internal/enums/sysrole"
	"github.com/loveyourstack/lys-ref/internal/stores/core/corearraytype"
	"github.com/loveyourstack/lys-ref/internal/stores/core/coredefaultvalue"
	"github.com/loveyourstack/lys-ref/internal/stores/core/coremandatoryvalue"
	"github.com/loveyourstack/lys-ref/internal/stores/core/coreoptionalvalue"
	"github.com/loveyourstack/lys-ref/internal/stores/core/corevarianttype"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaign"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaignopt"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampperf"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys-ref/internal/stores/ecb/ecbcurr"
	"github.com/loveyourstack/lys-ref/internal/stores/ecb/ecbcurrmd"
	"github.com/loveyourstack/lys-ref/internal/stores/ecb/ecbxrperfnorm"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geocountry"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geoocean"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procflow"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procpoint"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procrun"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procstep"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procsteplink"
	"github.com/loveyourstack/lys-ref/internal/stores/publisher/pubauthor"
	"github.com/loveyourstack/lys-ref/internal/stores/publisher/pubauthorarch"
	"github.com/loveyourstack/lys-ref/internal/stores/publisher/pubbook"
	"github.com/loveyourstack/lys-ref/internal/stores/publisher/pubbookarch"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppcompany"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppemployee"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppprodcategory"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppproduct"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysnotification"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/loveyourstack/lys/lyspgmon/stores/lyspgauditupdate"
	"github.com/loveyourstack/lys/lyspgmon/stores/lyspgbloat"
	"github.com/loveyourstack/lys/lyspgmon/stores/lyspgquery"
	"github.com/loveyourstack/lys/lyspgmon/stores/lyspgsetting"
	"github.com/loveyourstack/lys/lyspgmon/stores/lyspgtablesize"
	"github.com/loveyourstack/lys/lyspgmon/stores/lyspgunusedidx"
)

// getRouter returns a mux providing the HTTP server's routes
func (srvApp *httpServerApplication) getRouter() http.Handler {

	// define env struct needed for route handlers
	apiEnv := lys.Env{
		Logger:      srvApp.Logger,
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

	r.NotFoundHandler = http.HandlerFunc(lys.NotFound())

	// public subrouter: use unauthenticated middleware
	pubR := r.NewRoute().Subrouter()
	pubR.Use(srvApp.limitUnauthed)
	pubR.Use(srvApp.logUnauthedRequest)

	pubR.HandleFunc("/login", srvApp.authLogin).Methods("POST")
	pubR.HandleFunc("/session-token-login", srvApp.authSessionTokenLogin).Methods("POST")

	// put all routes requiring auth behind "/a" for authed
	authedR := r.PathPrefix("/a").Subrouter()

	// define middleware for authed routes
	authedR.Use(srvApp.authenticate)
	authedR.Use(srvApp.limitAuthed)      // must come after authenticate
	authedR.Use(srvApp.logAuthedRequest) // must come after authenticate

	// logout
	authedR.HandleFunc("/logout", srvApp.authLogout).Methods("POST")

	// add subroutes into main router
	for _, subRoute := range srvApp.getSubRoutes(apiEnv) {
		subRouter := authedR.PathPrefix(subRoute.Url).Subrouter()
		_ = subRoute.RouteAdder(subRouter)
	}

	// apply CORS middleware to allow access to Vue app
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{srvApp.Config.UI.Url}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Accept", "Accept-Encoding", "Authorization", "Content-Length", "Content-Type", "X-CSRF-Token"}),
		handlers.ExposedHeaders([]string{"Content-Disposition"}),
		handlers.AllowCredentials(),
	)
	return (cors)(r)
}

// getSubRoutes returns all subroutes used by the server
func (srvApp *httpServerApplication) getSubRoutes(apiEnv lys.Env) []lys.SubRoute {

	return []lys.SubRoute{
		{Url: "/aws", RouteAdder: srvApp.awsRoutes(apiEnv)},
		{Url: "/core", RouteAdder: srvApp.coreRoutes(apiEnv)},
		{Url: "/digmark", RouteAdder: srvApp.digmarkRoutes(apiEnv)},
		{Url: "/ecb", RouteAdder: srvApp.ecbRoutes(apiEnv)},
		{Url: "/geo", RouteAdder: srvApp.geoRoutes(apiEnv)},
		{Url: "/maxmind", RouteAdder: srvApp.maxmindRoutes(apiEnv)},
		{Url: "/pg-monitor", RouteAdder: srvApp.pgMonRoutes(apiEnv)},
		{Url: "/process", RouteAdder: srvApp.procRoutes(apiEnv)},
		{Url: "/publisher", RouteAdder: srvApp.publisherRoutes(apiEnv)},
		{Url: "/supplier", RouteAdder: srvApp.supplierRoutes(apiEnv)},
		{Url: "/system", RouteAdder: srvApp.systemRoutes(apiEnv)},
		{Url: "/tech", RouteAdder: srvApp.techRoutes(apiEnv)},
		{Url: "/ws", RouteAdder: srvApp.wsRoutes()},
	}
}

func (srvApp *httpServerApplication) awsRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {

		// restrict data change routes to writer roles
		writeR := r.NewRoute().Subrouter()
		writeR.Use(authorizeRole(sysrole.Writer[:]))

		endpoint := "/api-calls"

		apiCallStore := awsapicall.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, apiCallStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, apiCallStore)).Methods("GET")

		endpoint = "/update-user-security-group-rules"
		writeR.HandleFunc(endpoint, srvApp.awsUpdateUserSecurityGroupRules).Methods("PATCH")

		return r
	}
}

func (srvApp *httpServerApplication) coreRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {
		schemaName := "core"

		// restrict data change routes to writer roles
		writeR := r.NewRoute().Subrouter()
		writeR.Use(authorizeRole(sysrole.Writer[:]))

		endpoint := "/array-types"

		arTypeStore := corearraytype.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, arTypeStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, arTypeStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, arTypeStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, arTypeStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, arTypeStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, arTypeStore)).Methods("DELETE")

		endpoint = "/default-values"

		defValueStore := coredefaultvalue.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, defValueStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, defValueStore)).Methods("GET")
		writeR.HandleFunc(endpoint+"/import", lys.Import(apiEnv, srvApp.Db, defValueStore)).Methods("POST")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, defValueStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, defValueStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, defValueStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, defValueStore)).Methods("DELETE")

		r.HandleFunc("/mandatory-enums", lys.GetEnumValues(apiEnv, srvApp.Db, schemaName, "mandatory_enum")).Methods("GET")

		endpoint = "/mandatory-values"

		mandValueStore := coremandatoryvalue.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, mandValueStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, mandValueStore)).Methods("GET")
		writeR.HandleFunc(endpoint+"/import", lys.Import(apiEnv, srvApp.Db, mandValueStore,
			lys.ImportValueRepl{StringJsonName: "c_table_name", Int64JsonName: "c_table_fk", MapFunc: geoocean.NameIdValueMap},
		)).Methods("POST")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, mandValueStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, mandValueStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, mandValueStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, mandValueStore)).Methods("DELETE")

		r.HandleFunc("/optional-enums", lys.GetEnumValues(apiEnv, srvApp.Db, schemaName, "optional_enum")).Methods("GET")

		endpoint = "/optional-values"

		optValueStore := coreoptionalvalue.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, optValueStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, optValueStore)).Methods("GET")
		writeR.HandleFunc(endpoint+"/import", lys.Import(apiEnv, srvApp.Db, optValueStore,
			lys.ImportValueRepl{StringJsonName: "c_table_name", Int64JsonName: "c_table_fk", MapFunc: geocountry.NameIdValueMap},
		)).Methods("POST")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, optValueStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, optValueStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, optValueStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, optValueStore)).Methods("DELETE")

		r.HandleFunc("/performance-periods", lys.GetEnumValues(apiEnv, srvApp.Db, schemaName, "performance_period")).Methods("GET")

		endpoint = "/variant-types"

		varTypeStore := corevarianttype.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, varTypeStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, varTypeStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, varTypeStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, varTypeStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, varTypeStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, varTypeStore)).Methods("DELETE")

		return r
	}
}

func (srvApp *httpServerApplication) digmarkRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {
		schemaName := "digmark"

		writeR := r.NewRoute().Subrouter()
		writeR.Use(authorizeRole(sysrole.Writer[:]))

		endpoint := "/campaigns"

		campStore := dmcampaign.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, campStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, campStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, campStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, campStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/active-by-ids", srvApp.dmPatchActiveByIds).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/budget-percent-by-ids", srvApp.dmPatchBudgetPercentByIds).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, campStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, campStore)).Methods("DELETE")

		endpoint = "/campaign-optimizer"

		campOptStore := dmcampaignopt.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint+"/aggregates", srvApp.dmGetCampaignOptAggregates).Methods("GET")
		r.HandleFunc(endpoint, lys.Get(apiEnv, campOptStore, &lys.GetOpts[dmcampaignopt.Model]{
			SetFuncUrlParamNames: campOptStore.GetSetFuncUrlParamNames(),
		})).Methods("GET")

		endpoint = "/campaign-performance"

		campPerfStore := dmcampperf.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, campPerfStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, campPerfStore)).Methods("GET")
		// no modification routes: assumed that campaign performance data is synced by background processes

		endpoint = "/campaign-performance-latest-summary"

		r.HandleFunc(endpoint, lys.GetSimple(apiEnv, campPerfStore.SelectLatestPerfSummary)).Methods("GET")

		r.HandleFunc("/managers", lys.GetEnumValues(apiEnv, srvApp.Db, schemaName, "manager")).Methods("GET")

		endpoint = "/manager-budgets"

		r.HandleFunc(endpoint, lys.GetSimple(apiEnv, campStore.SelectManagerBudgets)).Methods("GET")

		endpoint = "/mcp-query"

		r.HandleFunc(endpoint, srvApp.dmMcpQuery).Methods("POST")

		endpoint = "/verticals"

		vertStore := dmvertical.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, vertStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, vertStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, vertStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, vertStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, vertStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, vertStore)).Methods("DELETE")

		endpoint = "/vertical-budgets"

		r.HandleFunc(endpoint, lys.GetSimple(apiEnv, campStore.SelectVerticalBudgets)).Methods("GET")

		return r
	}
}

func (srvApp *httpServerApplication) ecbRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {
		//schemaName := "ecb"

		writeR := r.NewRoute().Subrouter()
		writeR.Use(authorizeRole(sysrole.Writer[:]))

		endpoint := "/api-calls"

		apiCallStore := ecbapicall.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, apiCallStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, apiCallStore)).Methods("GET")

		endpoint = "/currencies"

		currStore := ecbcurr.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, currStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, currStore)).Methods("GET")

		endpoint = "/currency-metadata"

		currMdStore := ecbcurrmd.Store{Db: srvApp.Db}
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, currMdStore)).Methods("PATCH")

		endpoint = "/exchange-rates"

		xrStore := ecbexchangerate.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, xrStore, nil)).Methods("GET")

		endpoint = "/xr-performance-normalized"

		xrPerfNormStore := ecbxrperfnorm.Store{Db: srvApp.Db}
		// override default max # results for this endpoint since it needs to return up to "Last 90 days" of daily data
		xrPerfNormStoreEnv := apiEnv
		xrPerfNormStoreEnv.GetOptions.MaxPerPage = 500
		r.HandleFunc(endpoint, lys.Get(xrPerfNormStoreEnv, xrPerfNormStore, nil)).Methods("GET")

		return r
	}
}

func (srvApp *httpServerApplication) geoRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {
		//schemaName := "geo"

		writeR := r.NewRoute().Subrouter()
		writeR.Use(authorizeRole(sysrole.Writer[:]))

		endpoint := "/countries"

		countryStore := geocountry.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, countryStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, countryStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, countryStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, countryStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, countryStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, countryStore)).Methods("DELETE")

		endpoint = "/maxmind-locations"

		mmlocationStore := mmlocation.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, mmlocationStore, nil)).Methods("GET")

		endpoint = "/maxmind-networks"

		mmnetworkStore := mmnetwork.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, mmnetworkStore, nil)).Methods("GET")

		endpoint = "/oceans"

		oceanStore := geoocean.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, oceanStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, oceanStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, oceanStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, oceanStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, oceanStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, oceanStore)).Methods("DELETE")

		return r
	}
}

func (srvApp *httpServerApplication) maxmindRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {
		//schemaName := "maxmind"

		endpoint := "/api-calls"

		apiCallStore := mmapicall.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, apiCallStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, apiCallStore)).Methods("GET")

		return r
	}
}

func (srvApp *httpServerApplication) pgMonRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {

		// restrict this subroute
		r.Use(authorizeRole([]sysrole.Enum{sysrole.Tech}))

		endpoint := "/bloat"

		bloatStore := lyspgbloat.Store{Db: srvApp.OwnerDb} // uses db owner
		r.HandleFunc(endpoint, lys.Get(apiEnv, bloatStore, nil)).Methods("GET")

		endpoint = "/database-size"

		r.HandleFunc(endpoint, lys.GetValue[string](apiEnv, srvApp.Db, lyspgdb.DbSizePrettyStmt)).Methods("GET")

		endpoint = "/queries"

		activityStore := lyspgquery.Store{Db: srvApp.OwnerDb} // uses db owner
		r.HandleFunc(endpoint, lys.Get(apiEnv, activityStore, nil)).Methods("GET")

		endpoint = "/settings"

		settingStore := lyspgsetting.Store{Db: srvApp.OwnerDb} // uses db owner
		r.HandleFunc(endpoint, lys.Get(apiEnv, settingStore, nil)).Methods("GET")

		endpoint = "/table-size"

		tableSizeStore := lyspgtablesize.Store{Db: srvApp.OwnerDb} // uses db owner
		r.HandleFunc(endpoint, lys.Get(apiEnv, tableSizeStore, nil)).Methods("GET")

		endpoint = "/unused-indexes"

		unusedIdxStore := lyspgunusedidx.Store{Db: srvApp.OwnerDb} // uses db owner
		r.HandleFunc(endpoint, lys.Get(apiEnv, unusedIdxStore, nil)).Methods("GET")

		endpoint = "/version"

		r.HandleFunc(endpoint, lys.GetValue[string](apiEnv, srvApp.Db, lyspgdb.VersionStmt)).Methods("GET")

		return r
	}
}

func (srvApp *httpServerApplication) procRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {

		writeR := r.NewRoute().Subrouter()
		writeR.Use(authorizeRole(sysrole.Writer[:]))

		endpoint := "/flows"

		flowStore := procflow.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, flowStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, flowStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, flowStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, flowStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, flowStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, flowStore)).Methods("DELETE")

		endpoint = "/points"

		pointStore := procpoint.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, pointStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, pointStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, pointStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, pointStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, pointStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, pointStore)).Methods("DELETE")

		endpoint = "/runs"

		runStore := procrun.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, runStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, runStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, runStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, runStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, runStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, runStore)).Methods("DELETE")

		endpoint = "/steps"

		stepStore := procstep.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, stepStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}/available-dependencies", srvApp.procGetStepAvailableDependencies).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, stepStore)).Methods("GET")
		writeR.HandleFunc(endpoint+"/{id}/run", srvApp.procRunStep).Methods("POST")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, stepStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/swap-display-order", srvApp.procSwapDisplayOrder).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, stepStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, stepStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, stepStore)).Methods("DELETE")

		endpoint = "/step-links"

		stepLinkStore := procsteplink.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, stepLinkStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, stepLinkStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, stepLinkStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, stepLinkStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, stepLinkStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, stepLinkStore)).Methods("DELETE")

		return r
	}
}

func (srvApp *httpServerApplication) publisherRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {
		//schemaName := "publisher"

		writeR := r.NewRoute().Subrouter()
		writeR.Use(authorizeRole(sysrole.Writer[:]))

		endpoint := "/authors"

		authorStore := pubauthor.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, authorStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, authorStore)).Methods("GET")
		writeR.HandleFunc(endpoint+"/{id}/restore", lys.Restore(apiEnv, srvApp.Db, authorStore)).Methods("POST")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, authorStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, authorStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, authorStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}/archive", lys.Archive(apiEnv, srvApp.Db, authorStore)).Methods("DELETE")

		endpoint = "/authors-archived"

		authorArchStore := pubauthorarch.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, authorArchStore, nil)).Methods("GET")

		endpoint = "/books"

		bookStore := pubbook.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, bookStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, bookStore)).Methods("GET")
		writeR.HandleFunc(endpoint+"/{id}/restore", lys.Restore(apiEnv, srvApp.Db, bookStore)).Methods("POST")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, bookStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, bookStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, bookStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}/archive", lys.Archive(apiEnv, srvApp.Db, bookStore)).Methods("DELETE")

		endpoint = "/books-archived"

		bookArchStore := pubbookarch.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, bookArchStore, nil)).Methods("GET")

		return r
	}
}

func (srvApp *httpServerApplication) supplierRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {
		//schemaName := "supplier"

		writeR := r.NewRoute().Subrouter()
		writeR.Use(authorizeRole(sysrole.Writer[:]))

		endpoint := "/companies"

		compStore := suppcompany.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, compStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, compStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, compStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, compStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, compStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, compStore)).Methods("DELETE")

		endpoint = "/employees"

		employeeStore := suppemployee.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, employeeStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, employeeStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, employeeStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, employeeStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, employeeStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, employeeStore)).Methods("DELETE")

		endpoint = "/product-categories"

		productCatStore := suppprodcategory.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, productCatStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, productCatStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, productCatStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, productCatStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, productCatStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, productCatStore)).Methods("DELETE")

		endpoint = "/products"

		productStore := suppproduct.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, productStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, productStore)).Methods("GET")
		writeR.HandleFunc(endpoint, lys.Post(apiEnv, productStore)).Methods("POST")
		writeR.HandleFunc(endpoint+"/{id}", lys.Put(apiEnv, productStore)).Methods("PUT")
		writeR.HandleFunc(endpoint+"/{id}", lys.Patch(apiEnv, productStore)).Methods("PATCH")
		writeR.HandleFunc(endpoint+"/{id}", lys.Delete(apiEnv, productStore)).Methods("DELETE")

		return r
	}
}

func (srvApp *httpServerApplication) systemRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {
		//schemaName := "system"

		endpoint := "/audit-updates"

		auditUpdateStore := lyspgauditupdate.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, auditUpdateStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, auditUpdateStore)).Methods("GET")

		endpoint = "/notifications"

		notsStore := sysnotification.Store{Db: srvApp.Db}

		// default select: override Get to only return notifications for the ctx user id
		r.HandleFunc(endpoint, lys.Get(apiEnv, notsStore, &lys.GetOpts[sysnotification.Model]{SelectFunc: notsStore.SelectOnlyUsers})).Methods("GET")

		r.HandleFunc(endpoint+"/unread-count", srvApp.sysGetUserUnreadNotificationCount).Methods("GET")
		r.HandleFunc(endpoint+"/add-fake", srvApp.sysAddFakeNotification).Methods("POST")
		r.HandleFunc(endpoint+"/set-all-read", srvApp.sysSetAllNotificationsToRead).Methods("PATCH")
		r.HandleFunc(endpoint+"/set-read", srvApp.sysSetNotificationsToRead(apiEnv)).Methods("PATCH")

		endpoint = "/ui-store-data"

		r.HandleFunc(endpoint, srvApp.sysGetUiStoreData).Methods("GET")

		return r
	}
}

func (srvApp *httpServerApplication) techRoutes(apiEnv lys.Env) lys.RouteAdderFunc {

	return func(r *mux.Router) *mux.Router {

		// restrict this subroute
		r.Use(authorizeRole([]sysrole.Enum{sysrole.Tech}))

		endpoint := "/blocked-ips"

		blockedIpStore := sysblockedip.Store{Db: srvApp.Db}
		r.HandleFunc(endpoint, lys.Get(apiEnv, blockedIpStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, blockedIpStore)).Methods("GET")

		endpoint = "/hub/status"

		r.HandleFunc(endpoint, srvApp.techHubStatus).Methods("GET")

		endpoint = "/login-attempts"

		r.HandleFunc(endpoint, srvApp.authGetLoginAttempts).Methods("GET")
		r.HandleFunc(endpoint+"/unblock-ip/{ip}", srvApp.authUnblockIp).Methods("POST")

		// for testing cancellation
		r.HandleFunc("/pg-sleep", lys.PgSleep(srvApp.Db, srvApp.Logger, 10, 0)).Methods("GET")

		endpoint = "/server-requests"

		srvReqStore := srvApp.SrvLogStore
		r.HandleFunc(endpoint, lys.Get(apiEnv, srvReqStore, nil)).Methods("GET")
		r.HandleFunc(endpoint+"/{id}", lys.GetById(apiEnv, srvReqStore)).Methods("GET")

		endpoint = "/sessions"

		r.HandleFunc(endpoint, srvApp.authGetSessions).Methods("GET")
		r.HandleFunc(endpoint+"/block-ip/{ip}", srvApp.authBlockSessionIp).Methods("POST")

		return r
	}
}

func (srvApp *httpServerApplication) wsRoutes() lys.RouteAdderFunc {
	return func(r *mux.Router) *mux.Router {

		endpoint := "/notifications/register" // needs ws:// protocol

		r.HandleFunc(endpoint, srvApp.wsNotificationsRegister).Methods("GET")

		return r
	}
}
