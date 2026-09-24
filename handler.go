package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func readinessHandler(rs http.ResponseWriter, req *http.Request) {
	rs.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rs.WriteHeader(200)
	rs.Write([]byte("OK"))
}

// ApiConfig Method
func (cfg *apiConfig) fileServerHitsHandler(rs http.ResponseWriter, req *http.Request) {
	rs.Header().Set("Content-Type", "text/html")
	rs.WriteHeader(200)
	body := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
	`, cfg.fileserverHits.Load())
	rs.Write([]byte(body))
}

func (cfg *apiConfig) resetFileServerHitsHandler(rs http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits = atomic.Int32{}
	rs.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rs.WriteHeader(200)
	rs.Write([]byte("OK"))
}
