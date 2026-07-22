-- name: UpdateAssetNameByID :execrows
UPDATE asset
SET name = ?
WHERE id = ?;

-- name: UpdateAssetWarrantyByID :execrows
UPDATE asset
SET warranty_expiry = ?
WHERE id = ?;

-- name: UpdateAssetStatusByID :execrows
UPDATE asset
SET status = ?
WHERE id = ?;

-- name: UpdateAssetEOLByID :execrows
UPDATE asset
SET end_of_life = ?
WHERE id = ?;

-- name: UpdateAssetFileByID :execrows
UPDATE asset_file SET
  content_hash = ?,
  mime_type = ?,
  original_filename = ?,
  category = ?
WHERE asset_id = ?;

-- name: UpdateAssetByID :execrows
UPDATE asset
SET
  name = ?,
  warranty_expiry = ?,
  status = ?,
  end_of_life = ?
WHERE id = ?;
