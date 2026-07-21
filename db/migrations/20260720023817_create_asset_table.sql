-- +goose Up
pragma foreign_keys = on
;

CREATE TABLE asset
(
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  warranty_expiry NUMERIC,
  status TEXT,
  end_of_life NUMERIC
);

CREATE TABLE asset_file
(
  content_hash TEXT PRIMARY KEY,
  mime_type TEXT NOT NULL,
  original_filename TEXT NOT NULL,
  category TEXT NOT NULL,
  asset_id TEXT NOT NULL,
  FOREIGN KEY(asset_id) REFERENCES asset(asset_id)
);


-- +goose Down
DROP TABLE asset;

DROP TABLE asset_file;
