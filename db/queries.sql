-- name: GetUser :one
SELECT id, username, email, password_hash, created_at
FROM users 
WHERE username = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, username, email, password_hash, created_at;