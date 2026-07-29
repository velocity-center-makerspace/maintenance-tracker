-- name: UpdateAssetNameByID :execrows
UPDATE assets
SET name = ?
WHERE id = ?;

-- name: UpdateAssetWarrantyByID :execrows
UPDATE assets
SET warranty_expiry = ?
WHERE id = ?;

-- name: UpdateAssetAvailabilityByID :execrows
UPDATE assets
SET availability = ?
WHERE id = ?;

-- name: UpdateAssetNeedsAttentionByID :execrows
UPDATE assets
SET attention_needed = ?
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
  availability = ?,
  attention_needed = ?,
  end_of_life = ?
WHERE id = ?;

