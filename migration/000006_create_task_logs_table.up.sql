CREATE TABLE task_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    log_date DATE NOT NULL,
    rating VARCHAR(20) NOT NULL,
    score SMALLINT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_task_logs_task_id_log_date ON task_logs (task_id, log_date);
CREATE INDEX idx_task_logs_user_id_log_date ON task_logs (user_id, log_date);
