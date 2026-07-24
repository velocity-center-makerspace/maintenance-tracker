-- name: ReadTaskByID :one
SELECT sql.embed(tasks), sqlc.embed(task_recurrences)
FROM tasks
INNER JOIN task_recurrences
  ON tasks.id = task_recurrences.task_id
WHERE tasks.id = ?;

-- name: ListTasksAfterFirst :many
SELECT sqlc.embed(tasks), sqlc.embed(task_recurrences)
FROM tasks
INNER JOIN task_recurrences
  ON tasks.id = task_recurrences.task_id
ORDER BY tasks.id
LIMIT ?;

-- name: ListTasksAfterCursor :many
SELECT sqlc.embed(tasks), sqlc.embed(task_recurrences)
FROM tasks
INNER JOIN task_recurrences
  ON tasks.id = task_recurrences.task_id
WHERE tasks.id > ?
ORDER BY tasks.id
LIMIT ?;

-- name: ListAllTasks :many
SELECT sqlc.embed(tasks), sqlc.embed(task_recurrences)
FROM tasks
INNER JOIN task_recurrences
  ON tasks.id = task_recurrences.task_id
ORDER BY tasks.id;

-- name: ListAllTaskIDs :many
SELECT id FROM tasks;

-- name: ReadCompletionLog :one
SELECT * FROM completion_logs
WHERE id = ?;

-- name: ListAllCompletionLogs :many
SELECT * FROM completion_logs;

-- name: ListCompletionLogsAfterFirst :many
SELECT * FROM completion_logs
ORDER BY id
LIMIT ?;

-- name: ListCompletionLogsAfterCursor :many
SELECT * FROM completion_logs
WHERE id > ?
ORDER BY id
LIMIT ?;
