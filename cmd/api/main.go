package main

import (
	"log"
	"net/http"

	"github.com/yudan-glitch/twitter-backend/internal/domain"
	"github.com/yudan-glitch/twitter-backend/internal/handlers"
	"github.com/yudan-glitch/twitter-backend/internal/storage/mock"
)

func main() {
	// 1. Have yet to implement a real database. So, for now create an
	// instance of the mock database in memory and seed it with
	// temporary users for testing.
	runtimeStore := &mock.MockUserStore{
		Users: map[string]domain.User{
			"susan":  {Username: "alice", Email: "alice@example.com"},
			"jessie": {Username: "jessie", Email: "jessie@example.com"},
		},
	}

	// 2. Setup standard library multiplexer router
	mux := http.NewServeMux()

	// 3. Inject mock database dependency into the handler.
	mux.HandleFunc("GET /api/v1/users/{name}", handlers.HandleGetSpecificUser(runtimeStore))

	// 4. Start the live local server
	log.Println("Server is running on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
