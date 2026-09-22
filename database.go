package main

import (
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func ConectarBanco() (*sql.DB, error) {
	// no docker nao tem .env, as variaveis vem do docker-compose. entao se nao achar o arquivo segue o baile.
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
