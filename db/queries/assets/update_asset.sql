-- name: UpdateAssetNameByID :execrows
UPDATE assets
SET name = ?
WHERE id = ?;

-- name: UpdateAssetWarrantyByID :execrows
UPDATE assets
SET warranty_expiry = ?
WHERE id = ?;

-- name: UpdateAssetStatusByID :execrows
UPDATE assets
SET status = ?
WHERE id = ?;

-- name: UpdateAssetEOLByID :execrows
UPDATE assets
SET end_of_life = ?
WHERE id = ?;

-- name: UpdateAssetByID :execrows
UPDATE assets
SET
  name = ?,
  warranty_expiry = ?,
  status = ?,
  end_of_life = ?
WHERE id = ?;

