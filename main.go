package main

import (
	"net/http"
	"sync/atomic"
)

//Routing libraries; Gorilla Mux and Chi

//Stateful Handler
//To keep track of something

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	var srv http.Server
	mux := http.NewServeMux()

	srv.Handler = mux
	srv.Addr = ":8080"

	fileServerHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))

	apiCfg := apiConfig{}

	//NON api
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	//APIs
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.fileServerHitsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetFileServerHitsHandler)

	srv.ListenAndServe()

}
