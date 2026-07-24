-- name: CreateAsset :execrows
INSERT INTO assets
(id, name, warranty_expiry, status, end_of_life)
VALUES (?, ?, ?, ?, ?);

-- name: CreateAssetFile :execrows
INSERT INTO asset_files
(content_hash, mime_type, original_filename, category, asset_id)
VALUES (?, ?, ?, ?, ?);

