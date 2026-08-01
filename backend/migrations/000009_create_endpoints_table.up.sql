CREATE TABLE endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    path VARCHAR(500) NOT NULL,
    method VARCHAR(10) NOT NULL DEFAULT 'GET',
    version VARCHAR(20) NOT NULL DEFAULT 'v1',
    is_active BOOLEAN DEFAULT true,
    db_type VARCHAR(50) NOT NULL,
    schema VARCHAR(100) DEFAULT 'public',
    table_name VARCHAR(255),
    func_name VARCHAR(255),
    params JSONB,
    operations JSONB,
    security_policy JSONB,
    auth_header VARCHAR(100) DEFAULT 'Authorization',
    param_headers JSONB,
    body_mapping JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (collection_id, path, method, version)
);

CREATE INDEX idx_endpoints_collection_id ON endpoints(collection_id);
CREATE INDEX idx_endpoints_path ON endpoints(path);
