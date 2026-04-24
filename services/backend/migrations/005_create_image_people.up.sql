CREATE TABLE image_people (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
  name     TEXT NOT NULL
);

CREATE INDEX ON image_people (name);
CREATE UNIQUE INDEX ON image_people (image_id, name);
