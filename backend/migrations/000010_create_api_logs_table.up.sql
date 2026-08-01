CREATE TABLE api_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL,
    workspace_id UUID,
    user_id UUID,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(1000) NOT NULL,
    status_code INTEGER NOT NULL,
    latency_ms BIGINT NOT NULL,
    log_level VARCHAR(20) NOT NULL,
    message TEXT,
    meta JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_api_logs_request_id ON api_logs(request_id);
CREATE INDEX idx_api_logs_workspace_id ON api_logs(workspace_id);
CREATE INDEX idx_api_logs_user_id ON api_logs(user_id);
CREATE INDEX idx_api_logs_status_code ON api_logs(status_code);
CREATE INDEX idx_api_logs_created_at ON api_logs(created_at);
