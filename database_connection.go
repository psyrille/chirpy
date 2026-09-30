package main

import (
	"database/sql"

	"github.com/psyrille/chirpy/internal/database"
)

func connectdb(dbUrl string, cfg *apiConfig) error {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return err
	}

	cfg.db = *database.New(db)
	return nil
}
