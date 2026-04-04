package database

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations() {
	log.Println("Running database migrations...")

	// Get all SQL files in migrations directory
	files, err := filepath.Glob("database/migrations/*.sql")
	if err != nil {
		log.Fatal("Failed to find migration files: ", err)
	}

	// Sort files by name to ensure order
	sort.Strings(files)

	for _, file := range files {
		log.Printf("Running migration: %s", filepath.Base(file))

		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal("Failed to read migration file: ", err)
		}

		query := string(content)
		_, err = DB.Exec(query)
		if err != nil {
			log.Fatal("Failed to execute migration: ", err)
		}
	}

	log.Println("All migrations completed successfully")
}

func MigrateDown() {
	log.Println("Rolling back migrations...")

	files, err := filepath.Glob("database/migrations/*.sql")
	if err != nil {
		log.Fatal("Failed to find migration files: ", err)
	}

	sort.Strings(files)

	// Run in reverse order
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		log.Printf("Rolling back: %s", filepath.Base(file))

		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal("Failed to read migration file: ", err)
		}

		query := string(content)
		// Look for DOWN section
		parts := strings.Split(query, "-- DOWN")
		if len(parts) < 2 {
			log.Printf("No DOWN section found for %s, skipping", filepath.Base(file))
			continue
		}

		_, err = DB.Exec(parts[1])
		if err != nil {
			log.Printf("Warning: Failed to rollback migration %s: %v", filepath.Base(file), err)
		}
	}

	log.Println("Rollback completed")
}
