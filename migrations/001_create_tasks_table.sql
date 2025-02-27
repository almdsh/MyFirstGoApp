CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    method VARCHAR(10) NOT NULL,
    url TEXT NOT NULL,
    headers JSONB,
    status VARCHAR(20) NOT NULL,
    http_status_code INTEGER,
    response_headers JSONB,
    response_length BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
