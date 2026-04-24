CREATE TABLE image_tags (
  image_id UUID REFERENCES images(id) ON DELETE CASCADE,
  tag_id   UUID REFERENCES tags(id)   ON DELETE CASCADE,
  PRIMARY KEY (image_id, tag_id)
);

CREATE INDEX ON image_tags (tag_id);
