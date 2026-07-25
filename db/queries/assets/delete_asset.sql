-- name: DeleteAsset :execrows
DELETE FROM assets
WHERE id = ?;

-- name: DeleteAssetFile :execrows
DELETE FROM asset_files
WHERE asset_id = ? AND content_hash = ?;

-- name: DeleteFile :execrows
DELETE FROM files
WHERE content_hash = ?;

