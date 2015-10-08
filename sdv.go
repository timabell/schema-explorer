package main

import (
	"net"
	"net/http"
	"net/http/fcgi"
)

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("<h1>bonjour!</h1>\n<p>Hello soapie</p>"))
}

func main() {
	listener, _ := net.Listen("tcp", "127.0.0.1:80")
	srv := new (FastCGIServer)
	fcgi.Serve(listener, srv)
}
