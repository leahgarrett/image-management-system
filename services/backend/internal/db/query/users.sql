-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: UpdateUserRole :one
UPDATE users SET role = @role WHERE id = @id RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users SET status = @status WHERE id = @id RETURNING *;
