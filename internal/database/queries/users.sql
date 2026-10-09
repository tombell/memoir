-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, display_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: LockUser :one
SELECT * FROM users WHERE id = $1 FOR UPDATE;

-- name: VerifyUserEmail :exec
UPDATE users SET email_verified_at = COALESCE(email_verified_at, NOW()), updated = NOW()
WHERE id = $1;

-- name: SetUserPassword :exec
UPDATE users SET password_hash = $2, updated = NOW() WHERE id = $1;
