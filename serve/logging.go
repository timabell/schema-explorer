package serve

import (
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

// wrap handler, log requests as they pass through
func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static") {
			log.Printf("Request: '%v' | %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		} else {
			dump, err := httputil.DumpRequest(r, false)
			if err != nil {
				log.Println("couldn't dump request")
				panic(err)
			}
			log.Printf("Request: '%v' | %s", r.RemoteAddr, dump)
		}
		next.ServeHTTP(w, r)
	})
}
