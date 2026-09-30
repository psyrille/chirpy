package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/psyrille/chirpy/internal/auth"
	"github.com/psyrille/chirpy/internal/database"
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

func (cfg *apiConfig) resetHandler(rs http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		rs.WriteHeader(403)
		return
	}
	cfg.fileserverHits = atomic.Int32{}
	err := cfg.db.DeleteUsers(context.Background())
	if err != nil {
		rs.WriteHeader(500)
		rs.Write([]byte("There was an error in resetting"))
		return
	}
	rs.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rs.WriteHeader(200)
	rs.Write([]byte("OK"))
}

func (cfg *apiConfig) createUserHandler(rs http.ResponseWriter, req *http.Request) {
	var input struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}

	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		http.Error(rs, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Password == nil {
		http.Error(rs, "Password is required", http.StatusBadRequest)
		return
	}

	if input.Email == nil {
		http.Error(rs, "Email is required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(*input.Password)
	if err != nil {
		http.Error(rs, "Internal server error", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	user, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		ID:             uuid.New(),
		Email:          *input.Email,
		HashedPassword: hashedPassword,
		CreatedAt: sql.NullTime{
			Time:  now,
			Valid: true,
		},
		UpdatedAt: sql.NullTime{
			Time:  now,
			Valid: true,
		},
	})

	if err != nil {
		http.Error(rs, fmt.Sprintf("An error occured during the creation of the user: %v", err), http.StatusInternalServerError)
		return
	}

	response := User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		http.Error(rs, fmt.Sprintf("An error occured in parsing the user: %v", err), http.StatusInternalServerError)
		return
	}

	rs.Header().Set("Content-Type", "application/json")
	rs.WriteHeader(http.StatusCreated)
	rs.Write(responseData)

}

func (cfg *apiConfig) userLoginHandler(rw http.ResponseWriter, req *http.Request) {
	var input struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}

	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		http.Error(rw, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Email == nil {
		http.Error(rw, "Email is required", http.StatusBadRequest)
		return
	}

	if input.Password == nil {
		http.Error(rw, "Password is required", http.StatusBadRequest)
		return
	}

	user, err := cfg.db.CheckUser(req.Context(), *input.Email)
	if err != nil {
		http.Error(rw, fmt.Sprintf("1 Internal server error: %v %v", err, user.HashedPassword), http.StatusInternalServerError)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(rw, "User not found", http.StatusNotFound)
		return
	}

	match, err := auth.CheckPasswordHash(*input.Password, user.HashedPassword)
	if !match {
		http.Error(rw, "Incorrect email or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(rw, fmt.Sprintf("2 Internal server error: %v", err), http.StatusInternalServerError)
		return
	}

	jwtToken, _ := auth.MakeJWT(user.ID, cfg.secret)
	refreshToken := auth.MakeRefreshToken()
	nowDate := time.Now()

	dbRefreshToken, err := cfg.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		CreatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UserID: user.ID,
		ExpiresAt: sql.NullTime{
			Time:  nowDate.AddDate(0, 0, 60),
			Valid: true,
		},
		RevokedAt: sql.NullTime{},
	})
	if err != nil {
		http.Error(rw, fmt.Sprintf("3 Internal server error: %v", err), http.StatusInternalServerError)
		return
	}

	response := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt.Time,
		UpdatedAt:    user.UpdatedAt.Time,
		Email:        user.Email,
		Token:        jwtToken,
		RefreshToken: dbRefreshToken.Token,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		http.Error(rw, fmt.Sprintf("3 Internal server error: %v", err), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(responseData)
}

func (cfg *apiConfig) refreshHandler(rw http.ResponseWriter, req *http.Request) {
	type ReturnValue struct {
		Token string `json:"token"`
	}
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Bad Request: %v", err), http.StatusBadRequest)
		return
	}

	refreshTokenDb, err := cfg.db.GetRefreshToken(req.Context(), refreshToken)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(rw, "Refresh token not found", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	if refreshTokenDb.RevokedAt.Time.After(time.Now()) {
		http.Error(rw, "Refresh token expired", http.StatusUnauthorized)
		return
	}

	jwtToken, err := auth.MakeJWT(refreshTokenDb.UserID, cfg.secret)
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseData, err := json.Marshal(ReturnValue{Token: jwtToken})

	rw.WriteHeader(http.StatusOK)
	rw.Write(responseData)
}

func (cfg *apiConfig) revokeHandler(rw http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Bad Request: %v", err), http.StatusBadRequest)
		return
	}

	err = cfg.db.RevokeRefreshToken(req.Context(), database.RevokeRefreshTokenParams{
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		Token: refreshToken,
	})

	if err != nil {
		http.Error(rw, "Internal server error", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusNoContent)

}
