-- name: GetUser :one
SELECT id, username, email, password_hash, created_at
FROM user_account
WHERE username = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO user_account (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, username, email, password_hash, created_at;

-- name: GetUserByEmail :one
SELECT id, password_hash
FROM user_account
WHERE email = $1;