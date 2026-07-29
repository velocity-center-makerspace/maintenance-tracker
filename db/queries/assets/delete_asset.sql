-- name: DeleteAssetByID :execrows
DELETE FROM assets
WHERE id = ?;

-- name: DeleteAssetFileByAssetID :execrows
DELETE FROM asset_files
WHERE asset_id = ?;

-- name: DeleteFileByContentHash :execrows
DELETE FROM files
WHERE content_hash = ?;

