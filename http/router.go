package http

import (
	"bitbucket.org/timabell/sql-data-viewer/resources"
	"github.com/gorilla/mux"
	"net/http"
)

func Router() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.FileServer(http.Dir(resources.BasePath)))
	r.HandleFunc("/setup", SetupHandler)
	r.HandleFunc("/setup/{driver}", SetupDriverHandler)
	r.HandleFunc("/setup/{driver}/run", SetupDriverPostHandler)
	r.HandleFunc("/", DatabaseListHandler)
	r.HandleFunc("/{database}/", TableListHandler)
	r.HandleFunc("/{database}/table-trail", TableTrailHandler)
	r.HandleFunc("/{database}/table-trail/clear", ClearTableTrailHandler)
	r.HandleFunc("/{database}/tables/{tableName}", TableInfoHandler)
	r.HandleFunc("/{database}/tables/{tableName}/data", TableDataHandler)
	r.HandleFunc("/{database}/tables/{tableName}/analyse-data", AnalyseTableHandler)
	r.Use(loggingHandler)
	return r
}
