-- name: CreateTask :execrows
INSERT INTO tasks
(id, name, last_completed, next_date, start_date, end_date, asset_id)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: CreateTaskRecurrence :execrows
INSERT INTO task_recurrences
(id, amount, duration_type, task_id)
VALUES (?, ?, ?, ?);

-- name: CreateCompletionLog :execrows
INSERT INTO completion_logs
(id, completed_at, task_id)
VALUES (?, ?, ?);

