CREATE TABLE workspace_teams (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'viewer',
    PRIMARY KEY (workspace_id, team_id)
);

CREATE INDEX idx_workspace_teams_team_id ON workspace_teams(team_id);
