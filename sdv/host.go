package sdv

import (
	"fmt"
	"log"
	"net/http"
)

func Serve(handler func(http.ResponseWriter, *http.Request), port int) {
	// todo: use multiple handlers properly
	http.HandleFunc("/", handler)
	listenOn := fmt.Sprintf("localhost:%d", port)
	log.Printf("Starting server on http://%s/ - Press Ctrl-C to kill server.\n", listenOn)
	log.Fatal(http.ListenAndServe(listenOn, nil))
	log.Panic("http.ListenAndServe didn't block")
}
