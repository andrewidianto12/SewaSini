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
