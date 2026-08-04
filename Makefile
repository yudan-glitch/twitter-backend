# Load environment variables from .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: test

# Automatically migrates the test DB up, runs tests
test:
	migrate -path db/migrations -database "$(TEST_DATABASE_URL)" up
	go test ./internal/storage/postgres

# Automatically migrates the test DB up, runs tests (verbose flag)
test-v:
	migrate -path db/migrations -database "$(TEST_DATABASE_URL)" up
	go test -v -coverpkg=./internal/storage/postgres ./internal/storage/postgres

# These targets are dedicated to the main development database
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down

# Seed development database
seed:
	go run cmd/seed/main.go

# Run server
run:
	go run cmd/api/main.go
