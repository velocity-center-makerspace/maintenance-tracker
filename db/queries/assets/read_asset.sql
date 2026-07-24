-- name: ReadAsset :one
SELECT sqlc.embed(assets), sqlc.embed(asset_files)
FROM assets
INNER JOIN asset_files
  ON assets.id = asset_files.asset_id
WHERE assets.id = ?;

-- name: ListAssetsAfterFirst :many
SELECT sqlc.embed(assets), sqlc.embed(asset_files)
FROM assets
INNER JOIN asset_files
  ON assets.id = asset_files.asset_id
ORDER BY assets.id
LIMIT ?;

-- name: ListAssetsAfterCursor :many
SELECT sqlc.embed(assets), sqlc.embed(asset_files)
FROM assets
INNER JOIN asset_files
  ON assets.id = asset_files.asset_id
WHERE assets.id > ?
ORDER BY assets.id
LIMIT ?;

-- name: ListAllAssets :many
SELECT sqlc.embed(assets), sqlc.embed(asset_files)
FROM assets
INNER JOIN asset_files
  ON assets.id = asset_files.asset_id
ORDER BY assets.id;

-- name: ListAllAssetIDs :many
SELECT id FROM assets;
