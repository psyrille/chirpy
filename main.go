package main

import (
	"net/http"
)

func main() {
	var srv http.Server
	mux := http.NewServeMux()

	srv.Handler = mux
	srv.Addr = ":8080"

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", readinessHandler)

	srv.ListenAndServe()

}

func readinessHandler(rs http.ResponseWriter, req *http.Request) {
	rs.Header().Set("Content-Type", "text/plain;charset=utf-8")
	rs.WriteHeader(200)
	rs.Write([]byte("OK"))
}
