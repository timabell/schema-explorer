package main

import (
	"fmt"
	"net/http"
)

func handler(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(resp, "<h1>bonjour!</h1>\n<p>Hello soapie</p>")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
