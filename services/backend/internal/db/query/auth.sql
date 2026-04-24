-- name: CreateUser :one
INSERT INTO users (email, name, role, status, invited_by)
VALUES (@email, @name, @role, @status, @invited_by)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = @email LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = @id LIMIT 1;

-- name: UpdateUserLastLogin :exec
UPDATE users SET last_login_at = now() WHERE id = @id;

-- name: CreateMagicLinkToken :one
INSERT INTO magic_link_tokens (user_id, token_hash, expires_at)
VALUES (@user_id, @token_hash, @expires_at)
RETURNING *;

-- name: GetMagicLinkTokenByHash :one
SELECT * FROM magic_link_tokens WHERE token_hash = @token_hash LIMIT 1;

-- name: MarkTokenUsed :exec
UPDATE magic_link_tokens SET used_at = now() WHERE id = @id;
