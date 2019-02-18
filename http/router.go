package http

import (
	"github.com/gorilla/mux"
	"net/http"
)

func Router() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.FileServer(http.Dir(".")))
	r.HandleFunc("/", TableListHandler)
	r.HandleFunc("/table-trail", TableTrailHandler)
	r.HandleFunc("/table-trail/clear", ClearTableTrailHandler)
	r.HandleFunc("/tables/{tableName}", TableInfoHandler)
	r.HandleFunc("/tables/{tableName}/data", TableDataHandler)
	r.HandleFunc("/tables/{tableName}/analyse-data", AnalyseTableHandler)
	r.Use(loggingHandler)
	return r
}