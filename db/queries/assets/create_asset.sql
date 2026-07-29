-- name: CreateAsset :execrows
INSERT INTO assets
(id, name, warranty_expiry, availability, attention_needed, end_of_life)
VALUES (?, ?, ?, ?, ?, ?);

-- name: CreateAssetFile :execrows
INSERT INTO asset_files
(content_hash, asset_id, original_filename)
VALUES (?, ?, ?);

-- name: CreateFile :execrows
INSERT INTO files
(content_hash, mime_type, path)
VALUES (?, ?, ?);

