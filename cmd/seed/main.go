package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Populate development database
func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("CRITICAL: DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 1. Array of diverse mock accounts to seed the main environment
	usersToSeed := []struct {
		username string
		email    string
	}{
		{"susan", "susan@example.com"},
		{"jessie", "jessie@example.com"},
		{"alex", "alex@example.com"},
		{"alice", "alice@example.com"},
		{"elena", "elena@example.com"},
	}

	log.Println("Starting core development database seeding...")

	// 2. Loop through and execute parameterized inserts
	for _, u := range usersToSeed {
		query := `
			INSERT INTO users (username, email) 
			VALUES ($1, $2) 
			ON CONFLICT (username) DO NOTHING
		`
		_, err := db.Exec(query, u.username, u.email)
		if err != nil {
			log.Printf("Warning: Failed to seed user %s: %v", u.username, err)
			continue
		}
	}

	fmt.Println("==================================================")
	fmt.Println(" SUCCESS: Development database seeding complete! ")
	fmt.Println("==================================================")
}
