-- name: CreateTag :one
INSERT INTO tags (name, created_by)
VALUES (@name, @created_by)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name;

-- name: SearchTags :many
SELECT * FROM tags
WHERE name ILIKE '%' || sqlc.arg('query')::text || '%'
ORDER BY usage_count DESC
LIMIT 10;

-- name: AddImageTag :exec
INSERT INTO image_tags (image_id, tag_id) VALUES (@image_id, @tag_id)
ON CONFLICT DO NOTHING;

-- name: RemoveImageTag :exec
DELETE FROM image_tags WHERE image_id = @image_id AND tag_id = @tag_id;

-- name: ListImageTags :many
SELECT t.* FROM tags t
JOIN image_tags it ON it.tag_id = t.id
WHERE it.image_id = @image_id
ORDER BY t.name;

-- name: IncrementTagUsage :exec
UPDATE tags SET usage_count = usage_count + 1 WHERE id = @id;

-- name: DecrementTagUsage :exec
UPDATE tags SET usage_count = GREATEST(0, usage_count - 1) WHERE id = @id;
