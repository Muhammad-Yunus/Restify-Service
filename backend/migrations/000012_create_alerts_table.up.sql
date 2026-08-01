CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    workspace_id UUID NOT NULL,
    endpoint_id UUID,
    trigger VARCHAR(50) NOT NULL,
    threshold DOUBLE PRECISION NOT NULL,
    window_minutes INTEGER NOT NULL,
    actions JSONB,
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE alert_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    trigger VARCHAR(50) NOT NULL,
    current_value DOUBLE PRECISION NOT NULL,
    threshold DOUBLE PRECISION NOT NULL,
    message TEXT,
    notified BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_alert_rules_workspace_id ON alert_rules(workspace_id);
CREATE INDEX idx_alert_rules_endpoint_id ON alert_rules(endpoint_id);
CREATE INDEX idx_alert_events_rule_id ON alert_events(rule_id);
CREATE INDEX idx_alert_events_workspace_id ON alert_events(workspace_id);
