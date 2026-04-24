-- name: CreateImage :one
INSERT INTO images (
  image_id, original_filename, thumbnail_key, web_key, original_key,
  thumbnail_size, web_size, original_size, width, height, uploaded_by, exif
) VALUES (
  @image_id, @original_filename, @thumbnail_key, @web_key, @original_key,
  @thumbnail_size, @web_size, @original_size, @width, @height, @uploaded_by, @exif
)
RETURNING *;

-- name: GetImageByID :one
SELECT * FROM images WHERE id = @id LIMIT 1;

-- name: GetImageByImageID :one
SELECT * FROM images WHERE image_id = @image_id LIMIT 1;

-- name: ListImages :many
SELECT * FROM images
WHERE (sqlc.narg('occasion_category')::text IS NULL OR occasion_category = sqlc.narg('occasion_category')::text)
ORDER BY uploaded_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: UpdateImage :one
UPDATE images SET
  published         = COALESCE(sqlc.narg('published'), published),
  date_type         = COALESCE(sqlc.narg('date_type'), date_type),
  exact_date        = COALESCE(sqlc.narg('exact_date'), exact_date),
  start_date        = COALESCE(sqlc.narg('start_date'), start_date),
  end_date          = COALESCE(sqlc.narg('end_date'), end_date),
  occasion_category = COALESCE(sqlc.narg('occasion_category'), occasion_category),
  occasion_name     = COALESCE(sqlc.narg('occasion_name'), occasion_name)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteImage :exec
DELETE FROM images WHERE id = @id;

-- name: CreateImagePerson :one
INSERT INTO image_people (image_id, name) VALUES (@image_id, @name) RETURNING *;

-- name: DeleteImagePeople :exec
DELETE FROM image_people WHERE image_id = @image_id;

-- name: ListImagePeople :many
SELECT * FROM image_people WHERE image_id = @image_id ORDER BY name;
