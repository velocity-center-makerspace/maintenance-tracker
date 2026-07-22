-- name: ReadAsset :one
SELECT sqlc.embed(asset), sqlc.embed(asset_file)
FROM asset
INNER JOIN asset_file
  ON asset.id = asset_file.asset_id
WHERE asset.id = ?;

-- name: ListAssetsAfterFirst :many
SELECT sqlc.embed(asset), sqlc.embed(asset_file)
FROM asset
INNER JOIN asset_file
  ON asset.id = asset_file.asset_id
ORDER BY asset.id
LIMIT ?;

-- name: ListAssetsAfterCursor :many
SELECT sqlc.embed(asset), sqlc.embed(asset_file)
FROM asset
INNER JOIN asset_file
  ON asset.id = asset_file.asset_id
WHERE asset.id > ?
ORDER BY asset.id
LIMIT ?;

-- name: ListAllAssets :many
SELECT sqlc.embed(asset), sqlc.embed(asset_file)
FROM asset
INNER JOIN asset_file
  ON asset.id = asset_file.asset_id
ORDER BY asset.id;
