-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE assets
(
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  warranty_expiry TEXT,
  availability TEXT NOT NULL,
  attention_needed TEXT NOT NULL,
  end_of_life TEXT
);

CREATE TABLE asset_files
(
  asset_id TEXT NOT NULL,
  content_hash TEXT NOT NULL,
  original_filename TEXT NOT NULL,
  FOREIGN KEY (content_hash) REFERENCES files(content_hash),
  FOREIGN KEY (asset_id) REFERENCES assets(id)
);

CREATE TABLE files
(
  content_hash TEXT PRIMARY KEY,
  mime_type TEXT NOT NULL,
  path TEXT NOT NULL
);


-- +goose Down
DROP TABLE assets;

DROP TABLE asset_files;

DROP TABLE files;

