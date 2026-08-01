# Epic 10 — Team Service & Repository

**Goal:** Implement team repository, service, and HTTP handlers for team management.
**Dependencies:** Epic 04 (Team entity), Epic 05 (Repository interfaces), Epic 06 (DB adapter), Epic 09 (Workspace service)
**Commit:** `feat: implement team repository, service, and HTTP handlers`

---

## Step 10.01 — Team Repository Implementation

**Build:** Create `backend/internal/application/repository/team_repository.go`:

```go
package repository

import (
    "context"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
    "gorm.io/gorm"
)

// TeamRepositoryImpl implements the repository.TeamRepository interface.
type TeamRepositoryImpl struct {
    db *gorm.DB
}

// NewTeamRepository creates a new team repository.
func NewTeamRepository(db repository.DB) repository.TeamRepository {
    return &TeamRepositoryImpl{db: gormDB(db)}
}

func (r *TeamRepositoryImpl) Create(ctx context.Context, team *entity.Team) error {
    if team.ID == uuid.Nil {
        team.ID = uuid.New()
    }
    if err := r.db.WithContext(ctx).Create(team).Error; err != nil {
        return fmt.Errorf("create team: %w", err)
    }
    return nil
}

func (r *TeamRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Team, error) {
    var team entity.Team
    err := r.db.WithContext(ctx).Preload("Members.User.Roles").First(&team, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("find team %s: %w", id, err)
    }
    return &team, nil
}

func (r *TeamRepositoryImpl) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Team, error) {
    var teams []*entity.Team
    if err := r.db.WithContext(ctx).
        Preload("Members.User.Roles").
        Where("workspace_id = ?", workspaceID).
        Find(&teams).Error; err != nil {
        return nil, fmt.Errorf("list teams for workspace %s: %w", workspaceID, err)
    }
    return teams, nil
}

func (r *TeamRepositoryImpl) Update(ctx context.Context, team *entity.Team) error {
    if err := r.db.WithContext(ctx).Model(team).Update("name", team.Name).Error; err != nil {
        return fmt.Errorf("update team %s: %w", team.ID, err)
    }
    return nil
}

func (r *TeamRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
    result := r.db.WithContext(ctx).Delete(&entity.Team{}, "id = ?", id)
    if result.Error != nil {
        return fmt.Errorf("delete team %s: %w", id, result.Error)
    }
    if result.RowsAffected == 0 {
        return entity.ErrNotFound
    }
    return nil
}

func (r *TeamRepositoryImpl) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
    member := &entity.TeamMember{
        TeamID: teamID,
        UserID: userID,
        Role:   role,
    }
    if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
        return fmt.Errorf("add member to team: %w", err)
    }
    return nil
}

func (r *TeamRepositoryImpl) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
    result := r.db.WithContext(ctx).
        Where("team_id = ? AND user_id = ?", teamID, userID).
        Delete(&entity.TeamMember{})
    if result.Error != nil {
        return fmt.Errorf("remove member: %w", result.Error)
    }
    if result.RowsAffected == 0 {
        return entity.ErrNotFound
    }
    return nil
}

func (r *TeamRepositoryImpl) GetMember(ctx context.Context, teamID, userID uuid.UUID) (*entity.TeamMember, error) {
    var member entity.TeamMember
    err := r.db.WithContext(ctx).
        Preload("User.Roles").
        Where("team_id = ? AND user_id = ?", teamID, userID).
        First(&member).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("get team member: %w", err)
    }
    return &member, nil
}

func (r *TeamRepositoryImpl) ListMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.TeamMember, error) {
    var members []*entity.TeamMember
    if err := r.db.WithContext(ctx).
        Preload("User.Roles").
        Where("team_id = ?", teamID).
        Find(&members).Error; err != nil {
        return nil, fmt.Errorf("list team members: %w", err)
    }
    return members, nil
}

func (r *TeamRepositoryImpl) AssignWorkspace(ctx context.Context, teamID, workspaceID uuid.UUID, role entity.TeamWorkspaceRole) error {
    // Uses workspace_teams join table
    type WorkspaceTeam struct {
        TeamID      uuid.UUID `gorm:"column:team_id;primaryKey"`
        WorkspaceID uuid.UUID `gorm:"column:workspace_id;primaryKey"`
        Role        string    `gorm:"column:role;type:varchar(50)"`
    }
    wt := WorkspaceTeam{TeamID: teamID, WorkspaceID: workspaceID, Role: string(role)}
    if err := r.db.WithContext(ctx).Table("workspace_teams").
        Where("team_id = ? AND workspace_id = ?", teamID, workspaceID).
        First(&wt).Error; errors.Is(err, gorm.ErrRecordNotFound) {
        // Create new assignment
        if err := r.db.WithContext(ctx).Table("workspace_teams").Create(&wt).Error; err != nil {
            return fmt.Errorf("assign workspace to team: %w", err)
        }
    } else {
        // Update existing assignment
        if err := r.db.WithContext(ctx).Table("workspace_teams").
            Where("team_id = ? AND workspace_id = ?", teamID, workspaceID).
            Update("role", string(role)).Error; err != nil {
            return fmt.Errorf("update workspace assignment: %w", err)
        }
    }
    return nil
}

func (r *TeamRepositoryImpl) GetWorkspaceAccess(ctx context.Context, teamID uuid.UUID) ([]*entity.Workspace, error) {
    var workspaces []*entity.Workspace
    if err := r.db.WithContext(ctx).
        Table("workspace_teams").
        Select("workspaces.*").
        Joins("LEFT JOIN workspaces ON workspace_teams.workspace_id = workspaces.id").
        Where("workspace_teams.team_id = ?", teamID).
        Find(&workspaces).Error; err != nil {
        return nil, fmt.Errorf("get workspace access: %w", err)
    }
    return workspaces, nil
}
```

**Test cases:**
- [ ] Unit: `Create()` creates team with generated UUID
- [ ] Unit: `FindByID()` returns team with members
- [ ] Unit: `ListByWorkspace()` returns all teams in workspace
- [ ] Unit: `AddMember()` adds user to team
- [ ] Unit: `RemoveMember()` removes user from team
- [ ] Unit: `AssignWorkspace()` assigns team to workspace
- [ ] Integration: Full team lifecycle with test DB

---

## Step 10.02 — Team Service

**Build:** Create `backend/internal/application/service/team_service.go`:

```go
package service

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// TeamService implements the service.TeamService interface.
type TeamService struct {
    repo   repository.TeamRepository
    logger repository.Logger
}

// NewTeamService creates a new team service.
func NewTeamService(repo repository.TeamRepository, logger repository.Logger) repository.TeamService {
    return &TeamService{repo: repo, logger: logger}
}

func (s *TeamService) Create(ctx context.Context, name string, workspaceID uuid.UUID) (*entity.Team, error) {
    team := &entity.Team{
        Name:        name,
        WorkspaceID: workspaceID,
    }

    if err := team.Validate(); err != nil {
        return nil, fmt.Errorf("validate team: %w", err)
    }

    if err := s.repo.Create(ctx, team); err != nil {
        return nil, fmt.Errorf("create team: %w", err)
    }

    s.logger.Info(ctx, "team created", "team_id", team.ID, "workspace_id", workspaceID)
    return team, nil
}

func (s *TeamService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Team, error) {
    team, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get team %s: %w", id, err)
    }
    return team, nil
}

func (s *TeamService) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
    if err := s.repo.AddMember(ctx, teamID, userID, role); err != nil {
        return fmt.Errorf("add member to team: %w", err)
    }
    s.logger.Info(ctx, "team member added", "team_id", teamID, "user_id", userID)
    return nil
}

func (s *TeamService) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
    if err := s.repo.RemoveMember(ctx, teamID, userID); err != nil {
        return fmt.Errorf("remove member from team: %w", err)
    }
    s.logger.Info(ctx, "team member removed", "team_id", teamID, "user_id", userID)
    return nil
}

func (s *TeamService) ListMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.TeamMember, error) {
    return s.repo.ListMembers(ctx, teamID)
}

func (s *TeamService) AssignToWorkspace(ctx context.Context, teamID, workspaceID uuid.UUID, wsRole entity.TeamWorkspaceRole) error {
    if err := s.repo.AssignWorkspace(ctx, teamID, workspaceID, wsRole); err != nil {
        return fmt.Errorf("assign team to workspace: %w", err)
    }
    s.logger.Info(ctx, "team assigned to workspace", "team_id", teamID, "workspace_id", workspaceID, "role", wsRole)
    return nil
}
```

**Test cases:**
- [ ] Unit: `Create()` creates team in workspace
- [ ] Unit: `AddMember()` adds user with role
- [ ] Unit: `RemoveMember()` removes user
- [ ] Unit: `AssignToWorkspace()` assigns team with workspace role

---

## Step 10.03 — Team HTTP Handler

**Build:** Create `backend/internal/presentation/http/handler/team_handler.go`:

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/dto"
)

// TeamHandler handles team HTTP requests.
type TeamHandler struct {
    teamService service.TeamService
}

// NewTeamHandler creates a new team handler.
func NewTeamHandler(ts service.TeamService) *TeamHandler {
    return &TeamHandler{teamService: ts}
}

// Create handles POST /api/v1/workspaces/:ws_id/teams
func (h *TeamHandler) Create(c *gin.Context) {
    wsID, err := uuid.Parse(c.Param("ws_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid workspace ID", ""))
        return
    }

    var req struct {
        Name string `json:"name" validate:"required,max=255"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
        return
    }

    team, err := h.teamService.Create(c.Request.Context(), req.Name, wsID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusCreated, toTeamDTO(team))
}

// AddMember handles POST /api/v1/teams/:id/members
func (h *TeamHandler) AddMember(c *gin.Context) {
    teamID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid team ID", ""))
        return
    }

    var req struct {
        UserID string `json:"user_id" validate:"required,uuid"`
        Role   string `json:"role" validate:"oneof=member admin"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
        return
    }

    userID, err := uuid.Parse(req.UserID)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid user ID", ""))
        return
    }

    if err := h.teamService.AddMember(c.Request.Context(), teamID, userID, req.Role); err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "member added"})
}

// ListMembers handles GET /api/v1/teams/:id/members
func (h *TeamHandler) ListMembers(c *gin.Context) {
    teamID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid team ID", ""))
        return
    }

    members, err := h.teamService.ListMembers(c.Request.Context(), teamID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": toMemberListDTO(members)})
}

func toTeamDTO(team *entity.Team) gin.H {
    return gin.H{
        "id":           team.ID.String(),
        "name":         team.Name,
        "workspace_id": team.WorkspaceID.String(),
        "created_at":   team.CreatedAt,
    }
}

func toMemberListDTO(members []*entity.TeamMember) []gin.H {
    result := make([]gin.H, len(members))
    for i, m := range members {
        result[i] = gin.H{
            "id":       m.ID.String(),
            "user_id":  m.UserID.String(),
            "email":    m.User.Email,
            "full_name": m.User.FullName,
            "role":     m.Role,
            "joined_at": m.JoinedAt,
        }
    }
    return result
}
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: implement team repository, service, and HTTP handlers"
```
