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
WHERE (@occasion_category::text IS NULL OR occasion_category = @occasion_category)
ORDER BY uploaded_at DESC
LIMIT @lim OFFSET @off;

-- name: UpdateImage :one
UPDATE images SET
  published         = COALESCE(@published, published),
  date_type         = COALESCE(@date_type, date_type),
  exact_date        = COALESCE(@exact_date, exact_date),
  start_date        = COALESCE(@start_date, start_date),
  end_date          = COALESCE(@end_date, end_date),
  occasion_category = COALESCE(@occasion_category, occasion_category),
  occasion_name     = COALESCE(@occasion_name, occasion_name)
WHERE id = @id
RETURNING *;

-- name: DeleteImage :exec
DELETE FROM images WHERE id = @id;

-- name: CreateImagePerson :one
INSERT INTO image_people (image_id, name) VALUES (@image_id, @name) RETURNING *;

-- name: DeleteImagePeople :exec
DELETE FROM image_people WHERE image_id = @image_id;

-- name: ListImagePeople :many
SELECT * FROM image_people WHERE image_id = @image_id ORDER BY name;
