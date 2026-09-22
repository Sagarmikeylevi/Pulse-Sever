CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    recurrence_days JSONB,
    is_daily BOOLEAN NOT NULL DEFAULT false,
    target_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_tasks_user_id_type ON tasks (user_id, type) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_user_id_target_date ON tasks (user_id, target_date) WHERE type = 'check_in' AND deleted_at IS NULL;
CREATE INDEX idx_tasks_deleted_at ON tasks (deleted_at);
