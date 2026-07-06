CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL DEFAULT 'user',
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    word_id UUID NOT NULL,
    interval_days INT NOT NULL CHECK (interval_days IN (1, 3, 7, 30)),
    remind_at TIMESTAMP NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_schedules_due ON schedules(remind_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_schedules_user ON schedules(user_id);
