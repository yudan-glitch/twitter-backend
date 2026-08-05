package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/solid-state-dan/twitter-backend/internal/handlers"
	"github.com/solid-state-dan/twitter-backend/internal/storage/postgres"

	_ "github.com/lib/pq"
)

func main() {

	// 1. Parse configuration from standard environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("CRITICAL: DATABASE_URL environment variable is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback to safe default port
	}

	// 2. Open a connection pool to the database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to initialize database connection: %v", err)
	}
	// When opening a connection to a db, it is good practice to close it once done
	defer db.Close()

	// 3. Ping the database to confirm the network connection works properly (Health Check)
	if err := db.Ping(); err != nil {
		log.Fatalf("CRITICAL: Database connection is dead: %v", err)
	}
	log.Println("Database connection pool successfully verified.")

	// 4. Initialize the real PostgreSQL store implementation adapter
	userStore := postgres.NewPostgreSQLUserStore(db)

	// 5. Initialize standard library serve mux router
	mux := http.NewServeMux()

	// 6. Inject the production PostgreSQL store dependency into the standard library router
	mux.HandleFunc("GET /api/v1/users/{name}", handlers.HandleGetSpecificUser(userStore))
	mux.HandleFunc("POST /api/v1/users", handlers.HandleCreateUser(userStore))

	// 7. Start the live local server
	log.Println("Server is running on http://localhost:8080...")
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("CRITICAL: Server crashed: %v", err)
	}
}
