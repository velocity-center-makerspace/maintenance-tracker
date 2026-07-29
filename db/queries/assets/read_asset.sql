-- name: ReadAsset :one
SELECT sqlc.embed(assets), sqlc.embed(asset_files)
FROM assets
INNER JOIN asset_files
  ON assets.id = asset_files.asset_id
WHERE assets.id = ?;

-- name: ListAssetsAfterFirst :many
SELECT id, name
FROM assets
ORDER BY assets.id
LIMIT ?;

-- name: ListAssetsAfterCursor :many
SELECT id, name
FROM assets
WHERE assets.id > ?
ORDER BY assets.id
LIMIT ?;

-- name: ListAllAssets :many
SELECT id, name
FROM assets
ORDER BY assets.id;

-- name: ListAllAssetIDs :many
SELECT id FROM assets;

-- name: ReadContentHashByAssetID :many
SELECT content_hash
FROM asset_files
WHERE asset_id = ?;

-- name: CountAssetFileReferences :one
SELECT COUNT(*)
FROM asset_files
WHERE content_hash = ?;

-- name: ReadPathFromContentHash :one
SELECT path
FROM files
WHERE content_hash = ?;
