-- name: CreateAccountToken :exec
INSERT INTO account_tokens (token_hash, user_id, purpose, expires_at) VALUES ($1, $2, $3, $4);

-- name: GetAccountToken :one
SELECT user_id FROM account_tokens
WHERE token_hash = $1 AND purpose = $2 AND expires_at > NOW();

-- name: ConsumeAccountToken :one
DELETE FROM account_tokens
WHERE token_hash = $1 AND purpose = $2 AND expires_at > NOW()
RETURNING user_id;

-- name: DeleteUserAccountTokens :exec
DELETE FROM account_tokens WHERE user_id = $1 AND purpose = $2;

-- name: DeleteExpiredAccountTokens :exec
DELETE FROM account_tokens WHERE expires_at <= NOW();
