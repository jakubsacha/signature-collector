package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jakubsacha/signature-collector/models"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initialize the database
	config := models.DBConfig{
		Driver: "sqlite3",
		Name:   "local.db",
	}

	db, err := models.InitDB(config)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	// Read and execute migration files
	migrationFiles, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		log.Fatalf("Error finding migration files: %v", err)
	}

	for _, file := range migrationFiles {
		log.Printf("Applying migration: %s", file)

		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Error reading migration file %s: %v", file, err)
		}

		// Split the migration file into separate statements
		statements := strings.Split(string(content), ";")

		// Execute each statement
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if _, err := db.Exec(stmt); err != nil {
				log.Fatalf("Error executing migration %s: %v\nStatement: %s", file, err, stmt)
			}
		}

		fmt.Printf("✓ Applied migration: %s\n", filepath.Base(file))
	}

	fmt.Println("✓ All migrations completed successfully")
}
