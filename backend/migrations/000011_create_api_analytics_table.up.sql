CREATE TABLE api_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    endpoint_id UUID,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    labels JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_api_analytics_workspace_id ON api_analytics(workspace_id);
CREATE INDEX idx_api_analytics_endpoint_id ON api_analytics(endpoint_id);
CREATE INDEX idx_api_analytics_period_start ON api_analytics(period_start);
CREATE INDEX idx_api_analytics_metric ON api_analytics(workspace_id, metric_name, period_start);
