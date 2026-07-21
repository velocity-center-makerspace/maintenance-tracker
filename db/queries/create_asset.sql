-- name: CreateAsset :execrows
INSERT INTO asset 
  (id, name, warranty_expiry, status, end_of_life)
VALUES (?, ?, ?, ?, ?);

-- name: CreateAssetFile :execrows
INSERT INTO asset_file
  (content_hash, mime_type, original_filename, category, asset_id)
VALUES (?, ?, ?, ?, ?);
