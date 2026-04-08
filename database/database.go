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
	if err := ensureCoreSchemaCompatibility(); err != nil {
		log.Fatal("Failed to ensure core schema compatibility: ", err)
	}
	if err := ensureTransactionSchemaCompatibility(); err != nil {
		log.Fatal("Failed to ensure transactions schema compatibility: ", err)
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

func ensureTransactionSchemaCompatibility() error {
	const ensureTransactionsTable = `
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			booking_id INT NOT NULL,
			user_id INT NOT NULL REFERENCES users(id),
			amount BIGINT NOT NULL,
			payment_method VARCHAR(50) NOT NULL,
			transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			external_id VARCHAR(255) UNIQUE,
			xendit_id VARCHAR(255),
			last_webhook_id VARCHAR(255),
			payment_url TEXT,
			email_sent_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`
	if _, err := DB.Exec(ensureTransactionsTable); err != nil {
		return err
	}

	const ensureColumns = `
		ALTER TABLE transactions
		ADD COLUMN IF NOT EXISTS external_id VARCHAR(255),
		ADD COLUMN IF NOT EXISTS xendit_id VARCHAR(255),
		ADD COLUMN IF NOT EXISTS last_webhook_id VARCHAR(255),
		ADD COLUMN IF NOT EXISTS payment_url TEXT,
		ADD COLUMN IF NOT EXISTS email_sent_at TIMESTAMPTZ,
		ADD COLUMN IF NOT EXISTS transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	`
	if _, err := DB.Exec(ensureColumns); err != nil {
		return err
	}

	if _, err := DB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_external_id ON transactions(external_id)`); err != nil {
		return err
	}

	return nil
}

func ensureCoreSchemaCompatibility() error {
	const ensureKategoriTable = `
		CREATE TABLE IF NOT EXISTS kategori (
			id SERIAL PRIMARY KEY,
			nama_kategori VARCHAR(100) NOT NULL,
			deskripsi TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`
	if _, err := DB.Exec(ensureKategoriTable); err != nil {
		return err
	}

	const ensureRuanganTable = `
		CREATE TABLE IF NOT EXISTS ruangan (
			id SERIAL PRIMARY KEY,
			nama_ruangan VARCHAR(255) NOT NULL,
			kategori_id INT REFERENCES kategori(id) ON DELETE SET NULL,
			deskripsi TEXT,
			alamat VARCHAR(255),
			kota VARCHAR(100),
			kapasitas INT,
			harga_per_jam BIGINT,
			harga_per_hari BIGINT,
			stock_availability INT DEFAULT 1,
			fasilitas JSONB NOT NULL DEFAULT '[]'::jsonb,
			gambar JSONB NOT NULL DEFAULT '[]'::jsonb,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`
	if _, err := DB.Exec(ensureRuanganTable); err != nil {
		return err
	}

	const ensureBookingsTable = `
		CREATE TABLE IF NOT EXISTS bookings (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			ruangan_id INT NOT NULL REFERENCES ruangan(id) ON DELETE CASCADE,
			tanggal_mulai TIMESTAMPTZ NOT NULL,
			tanggal_selesai TIMESTAMPTZ NOT NULL,
			jumlah_peserta INT NOT NULL,
			total_harga BIGINT NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid',
			booking_code VARCHAR(50) UNIQUE NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`
	if _, err := DB.Exec(ensureBookingsTable); err != nil {
		return err
	}

	const ensureRuanganColumns = `
		ALTER TABLE ruangan
		ADD COLUMN IF NOT EXISTS fasilitas JSONB NOT NULL DEFAULT '[]'::jsonb,
		ADD COLUMN IF NOT EXISTS gambar JSONB NOT NULL DEFAULT '[]'::jsonb,
		ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
		ADD COLUMN IF NOT EXISTS stock_availability INT DEFAULT 1,
		ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	`
	if _, err := DB.Exec(ensureRuanganColumns); err != nil {
		return err
	}

	const ensureBookingColumns = `
		ALTER TABLE bookings
		ADD COLUMN IF NOT EXISTS total_harga BIGINT NOT NULL DEFAULT 0,
		ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending',
		ADD COLUMN IF NOT EXISTS payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid',
		ADD COLUMN IF NOT EXISTS booking_code VARCHAR(50),
		ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	`
	if _, err := DB.Exec(ensureBookingColumns); err != nil {
		return err
	}

	if _, err := DB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_code ON bookings(booking_code)`); err != nil {
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
