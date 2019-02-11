package main

import (
	"bitbucket.org/timabell/sql-data-viewer/host"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Http(t *testing.T) {
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	host.Router().ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("%d status for /", response.Code)
	}
}
