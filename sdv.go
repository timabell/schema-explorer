package main

import (
	"fmt"
	"net/http"
	"os"
)

var db string

func handler(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(resp, "<h1>bonjour!</h1>\n<p>Hello soapie</p>")
	fmt.Fprintf(resp, "<p>Connected to %s</p>", db)
}

func main() {
	db = os.Args[1]
	fmt.Printf("Connecting to db: %s\n", db)
	http.HandleFunc("/", handler)
	fmt.Println("Listening on http://localhost:8080/")
	fmt.Println("Press Ctrl-C to kill server")
	http.ListenAndServe(":8080", nil)
}
