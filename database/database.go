package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Prioritaskan PostgreSQL lokal; fallback ke konfigurasi Supabase bila kosong.
	host := getEnv("POSTGRES_HOST", os.Getenv("SUPABASE_HOST"))
	port := getEnv("POSTGRES_PORT", os.Getenv("SUPABASE_PORT"))
	user := getEnv("POSTGRES_USER", os.Getenv("SUPABASE_USER"))
	password := getEnv("POSTGRES_PASSWORD", os.Getenv("SUPABASE_PASSWORD"))
	dbname := getEnv("POSTGRES_DB", os.Getenv("SUPABASE_DB"))
	sslMode := os.Getenv("POSTGRES_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
		if os.Getenv("POSTGRES_HOST") == "" && os.Getenv("SUPABASE_HOST") != "" {
			sslMode = "require"
		}
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslMode,
	)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Test connection
	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	log.Println("Successfully connected to database")

	if err := ensureUserSchemaCompatibility(); err != nil {
		log.Fatal("Failed to ensure users schema compatibility: ", err)
	}
}

func ensureUserSchemaCompatibility() error {
	const ensureUsersTable = `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			nama_lengkap VARCHAR(255) NOT NULL,
			ttl DATE,
			no_hp VARCHAR(20),
			password VARCHAR(255) NOT NULL,
			role VARCHAR(10) NOT NULL DEFAULT 'user',
			otp_code VARCHAR(6),
			otp_expiry TIMESTAMP,
			is_verified BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`

	if _, err := DB.Exec(ensureUsersTable); err != nil {
		return err
	}

	const ensureColumns = `
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS otp_expiry TIMESTAMP,
		ADD COLUMN IF NOT EXISTS otp_code VARCHAR(6),
		ADD COLUMN IF NOT EXISTS is_verified BOOLEAN NOT NULL DEFAULT FALSE,
		ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	`

	if _, err := DB.Exec(ensureColumns); err != nil {
		return err
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallback
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}
