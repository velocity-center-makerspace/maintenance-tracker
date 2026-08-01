-- name: UpdateTaskNameByID :execrows
UPDATE tasks
SET name = ?
WHERE id = ?;

-- name: UpdateLastCompletedByID :execrows
UPDATE tasks
SET last_completed = ?
WHERE id = ?;

-- name: UpdateNextDateByID :execrows
UPDATE tasks
SET next_date = ?
WHERE id = ?;

-- name: UpdateStartDateByID :execrows
UPDATE tasks
SET start_date = ?
WHERE id = ?;

-- name: UpdateEndDateByID :execrows
UPDATE tasks
SET end_date = ?
WHERE id = ?;

-- name: UpdateAssetID :execrows
UPDATE tasks
SET asset_id = ?
WHERE id = ?;

-- name: UpdateTaskByID :execrows
UPDATE tasks
SET
  name = ?,
  last_completed = ?,
  next_date = ?,
  start_date = ?,
  end_date = ?,
  asset_id = ?
WHERE id = ?;

-- name: UpdateAmountByID :execrows
UPDATE task_recurrences
SET amount = ?
WHERE id = ?;

-- name: UpdateDurationTypeByID :execrows
UPDATE task_recurrences
SET duration_type = ?
WHERE id = ?;

-- name: UpdateTaskRecurrenceByID :execrows
UPDATE task_recurrences
SET
  amount = ?,
  duration_type = ?
WHERE id = ?;

