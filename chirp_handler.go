package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/psyrille/chirpy/internal/auth"
	"github.com/psyrille/chirpy/internal/database"
)

func validateChirpProfane(text string) (string, bool) {
	if len(text) > 140 {
		return "", false
	}

	texts := strings.Split(text, " ")
	profane_words := []string{"kerfuffle", "sharbert", "fornax"}

	for k, i_text := range texts {
		i_text = strings.ToLower(i_text)
		if slices.Contains(profane_words, i_text) {
			texts[k] = "****"
		}

	}

	return strings.Join(texts, " "), true

}

func (cfg *apiConfig) createChirps(rs http.ResponseWriter, req *http.Request) {
	var input struct {
		Body string `json:"body"`
	}

	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		http.Error(rs, "Invalid request body", http.StatusBadRequest)
		return
	}

	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(rs, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	userId, err := auth.ValidateJWT(bearerToken, cfg.secret)
	if err != nil {
		http.Error(rs, fmt.Sprintf("Token mismatch: %v", bearerToken), http.StatusUnauthorized)
		return
	}

	validatedBody, isValid := validateChirpProfane(input.Body)
	if !isValid {
		http.Error(rs, "Body too long", http.StatusBadRequest)
		return
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		Body:   validatedBody,
		UserID: userId,
	})

	if err != nil {
		http.Error(rs, fmt.Sprintf("There was an error in creating the chirp: %v", err), http.StatusInternalServerError)
		return
	}

	response := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt.Time,
		UpdatedAt: chirp.UpdatedAt.Time,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		http.Error(rs, fmt.Sprintf("There was an error during the parsing of the chirp data: %v", err), http.StatusInternalServerError)
		return
	}

	rs.WriteHeader(http.StatusCreated)
	rs.Write(responseData)
}

func (cfg *apiConfig) getChirps(rw http.ResponseWriter, req *http.Request) {
	var chirps []database.Chirp
	var err error
	sort := "asc"
	inputAuthorId := req.URL.Query().Get("author_id")
	if req.URL.Query().Get("sort") != "" {
		sort = req.URL.Query().Get("sort")
	}

	if inputAuthorId != "" {
		authorId, err := uuid.Parse(inputAuthorId)
		if err != nil {
			http.Error(rw, "Unable to parse uuid", http.StatusBadRequest)
			return
		}

		if sort == "desc" {
			chirps, err = cfg.db.GetChirpByAuthorIdOrderByDESC(req.Context(), authorId)
		} else {
			chirps, err = cfg.db.GetChirpByAuthorIdOrderByASC(req.Context(), authorId)
		}

	} else {
		chirps, err = cfg.db.GetChirps(req.Context(), sort)
		if err != nil {
			http.Error(rw, "There was a problem getting chirps", http.StatusInternalServerError)
			return
		}
	}

	response := []Chirp{}
	for _, c := range chirps {
		response = append(response, Chirp{
			ID:        c.ID,
			UserID:    c.UserID,
			CreatedAt: c.CreatedAt.Time,
			UpdatedAt: c.UpdatedAt.Time,
			Body:      c.Body,
		})
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		http.Error(rw, "There was a problem in parsing the chirps", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(responseData)

}

func (cfg *apiConfig) getChirpById(rw http.ResponseWriter, req *http.Request) {
	chirpId, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		http.Error(rw, "Not found", http.StatusNotFound)
	}
	if err != nil {
		http.Error(rw, fmt.Sprintf("Invalid UUID given: %v", err), http.StatusBadRequest)
		return
	}
	chirp, err := cfg.db.GetChirpById(req.Context(), chirpId)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(rw, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		http.Error(rw, fmt.Sprintf("There was a problem in getting the chirp: %v", err), http.StatusInternalServerError)
		return
	}

	response := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt.Time,
		UpdatedAt: chirp.UpdatedAt.Time,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		http.Error(rw, "There was a problem in parsing chirp", http.StatusInternalServerError)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(responseData)

}

func (cfg *apiConfig) deleteChirpHandler(rw http.ResponseWriter, req *http.Request) {

	chirpId, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		http.Error(rw, "Not found", http.StatusNotFound)
	}
	authToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userId, err := auth.ValidateJWT(authToken, cfg.secret)
	if err != nil {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chirp, err := cfg.db.GetChirpById(req.Context(), chirpId)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(rw, "Chirp not found", http.StatusNotFound)
		return
	}

	if chirp.UserID != userId {
		http.Error(rw, "Unauthorized", http.StatusForbidden)
		return
	}

	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = cfg.db.DeleteChirp(req.Context(), database.DeleteChirpParams{
		ID:     chirp.ID,
		UserID: userId,
	})

	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)

}
