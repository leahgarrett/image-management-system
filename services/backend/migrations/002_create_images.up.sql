CREATE TABLE images (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  image_id          TEXT NOT NULL UNIQUE,
  original_filename TEXT,
  thumbnail_key     TEXT,
  web_key           TEXT,
  original_key      TEXT,
  thumbnail_size    BIGINT,
  web_size          BIGINT,
  original_size     BIGINT,
  width             INTEGER,
  height            INTEGER,
  uploaded_by       UUID REFERENCES users(id),
  uploaded_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  published         BOOLEAN NOT NULL DEFAULT false,
  moderation_status TEXT NOT NULL DEFAULT 'pending'
    CHECK (moderation_status IN ('pending', 'approved', 'rejected')),
  date_type         TEXT CHECK (date_type IN ('exact', 'range', 'approximate')),
  exact_date        DATE,
  start_date        DATE,
  end_date          DATE,
  approx_year       INTEGER,
  approx_month      INTEGER,
  occasion_category TEXT CHECK (occasion_category IN (
    'birthday','wedding','graduation','holiday','vacation',
    'work_event','party','family_gathering','sports_event',
    'concert','conference','ceremony','casual','other'
  )),
  occasion_name     TEXT,
  exif              JSONB
);

CREATE INDEX ON images (uploaded_by);
CREATE INDEX ON images (uploaded_at DESC);
CREATE INDEX ON images (published);
CREATE INDEX ON images (date_type, exact_date);
