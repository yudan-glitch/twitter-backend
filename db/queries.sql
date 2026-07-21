-- name: GetUser :one
SELECT id, username, email, created_at
FROM users 
WHERE username = $1
LIMIT 1;