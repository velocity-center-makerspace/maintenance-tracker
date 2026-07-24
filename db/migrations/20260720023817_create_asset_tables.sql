-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE assets
(
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  warranty_expiry NUMERIC,
  status TEXT NOT NULL,
  end_of_life NUMERIC
);

CREATE TABLE asset_files
(
  content_hash TEXT PRIMARY KEY,
  mime_type TEXT NOT NULL,
  original_filename TEXT NOT NULL,
  category TEXT NOT NULL,
  asset_id TEXT NOT NULL,
  FOREIGN KEY (asset_id) REFERENCES assets(id)
);


-- +goose Down
DROP TABLE assets;

DROP TABLE asset_files;

