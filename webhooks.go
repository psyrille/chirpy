package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/psyrille/chirpy/internal/auth"
	"github.com/psyrille/chirpy/internal/database"
)

func (cfg *apiConfig) polka(rw http.ResponseWriter, req *http.Request) {
	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil {
		http.Error(rw, "", http.StatusUnauthorized)
		return
	}
	if apiKey != cfg.apiKey {
		http.Error(rw, "Invalid API key", http.StatusUnauthorized)
	}
	var input struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		http.Error(rw, "Invalid request", http.StatusBadRequest)
		return
	}

	if input.Event != "user.upgraded" {
		rw.WriteHeader(http.StatusNoContent)
		return
	}
	checkUser, err := cfg.db.GetUserData(req.Context(), input.Data.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(rw, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = cfg.db.UserUpgradeRed(req.Context(), database.UserUpgradeRedParams{
		ID: checkUser.ID,
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	})
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(204)

}
