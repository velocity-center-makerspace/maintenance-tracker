-- name: DeleteTaskByID :execrows
DELETE FROM tasks
WHERE id = ?;

-- name: DeleteTaskRecurrenceByID :execrows
DELETE FROM task_recurrences
WHERE id = ?;

-- name: DeleteCompletionLogByID :execrows
DELETE FROM completion_logs
WHERE id = ?;

