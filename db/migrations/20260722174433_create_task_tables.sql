-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE tasks
(
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  last_completed NUMERIC,
  next_date NUMERIC,
  start_date NUMERIC,
  end_date NUMERIC,
  asset_id TEXT NOT NULL,
  FOREIGN KEY (asset_id) REFERENCES assets(id)
);

CREATE TABLE task_recurrences
(
  id TEXT PRIMARY KEY,
  amount INTEGER CHECK (amount >= 1) NOT NULL,
  duration_type TEXT NOT NULL,
  task_id TEXT NOT NULL,
  FOREIGN KEY (task_id) REFERENCES tasks(id)
);

CREATE TABLE completion_logs
(
  id TEXT PRIMARY KEY,
  completed_at NUMERIC NOT NULL,
  task_id TEXT NOT NULL,
  FOREIGN KEY (task_id) REFERENCES tasks(id)
);

-- +goose Down
DROP TABLE tasks;

DROP TABLE task_recurrences;

