package main

import (
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/psyrille/chirpy/internal/database"
)

//Routing libraries; Gorilla Mux and Chi

//Stateful Handler
//To keep track of something

type apiConfig struct {
	fileserverHits atomic.Int32
	db             database.Queries
	platform       string
	secret         string
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
}

func main() {
	godotenv.Load()
	//Db Connection

	var srv http.Server
	mux := http.NewServeMux()

	srv.Handler = mux
	srv.Addr = ":8080"

	fileServerHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))

	apiCfg := apiConfig{
		platform: os.Getenv("PLATFORM"),
		secret:   os.Getenv("JWT_SECRET"),
	}
	connectdb(os.Getenv("DB_URL"), &apiCfg)

	//NON api
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	//APIs
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.fileServerHitsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)
	mux.HandleFunc("PUT /api/users", apiCfg.updateUserHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirps)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpById)
	mux.HandleFunc("POST /api/login", apiCfg.userLoginHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeHandler)

	srv.ListenAndServe()

}
