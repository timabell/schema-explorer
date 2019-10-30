package serve

import (
	"github.com/timabell/schema-explorer/resources"
	"github.com/gorilla/mux"
	"net/http"
)

func Router() (r *mux.Router) {
	r = mux.NewRouter()

	r.Use(loggingHandler)

	// static/*
	r.PathPrefix("/static/").Handler(http.FileServer(http.Dir(resources.BasePath)))

	// root
	r.HandleFunc("/", RootHandler)

	// setup/*
	setup := r.PathPrefix("/setup").Subrouter()
	setup.HandleFunc("", SetupHandler)
	setup.HandleFunc("/{driver}", SetupDriverHandler).Methods("GET")
	setup.HandleFunc("/{driver}", SetupDriverPostHandler).Methods("POST")

	// db list
	r.HandleFunc("/databases", DatabaseListHandler)

	/* database sub-route */
	database := r.PathPrefix("/{database}/").Subrouter()

	// Register all the per datbase routes twice:
	// once for when there is a /dbname/ prefix in the route:
	registerDatbaseRoutes(database, "multidb-")
	// and once for when the database choice is fixed (sqlite/azure/configured):
	registerDatbaseRoutes(r, "")

	return
}

func registerDatbaseRoutes(routerBase *mux.Router, namePrefix string) {
	// db info
	routerBase.HandleFunc("/", TableListHandler)
	// db/table/*
	tables := routerBase.PathPrefix("/tables/{tableName}").Subrouter()
	tables.HandleFunc("", TableInfoHandler).Name(namePrefix + "route-database-tables")
	tables.HandleFunc("/data", TableDataHandler)
	tables.HandleFunc("/analyse-data", AnalyseTableHandler)
	tables.HandleFunc("/description", TableDescriptionHandler).Methods("POST")
	tables.HandleFunc("/columns/{columnName}/description", ColumnDescriptionHandler).Methods("POST")
	trail := routerBase.PathPrefix("/table-trail").Subrouter()
	trail.HandleFunc("", TableTrailHandler)
	trail.HandleFunc("/clear", ClearTableTrailHandler)
}
