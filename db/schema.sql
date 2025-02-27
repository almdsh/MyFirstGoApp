CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    method VARCHAR(10) NOT NULL,
    url TEXT NOT NULL,
    headres JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    http_status_code INTEGER,
    response_headers JSONB,
    content_length BIGINT,
    creafted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
)

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
