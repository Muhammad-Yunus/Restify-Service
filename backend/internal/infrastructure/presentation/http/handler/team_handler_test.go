package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// fakeTeamService is a stub TeamService for handler tests.
type fakeTeamService struct {
	teams   []*entity.Team
	members []*entity.TeamMember
	listErr error
}

func (f *fakeTeamService) Create(_ context.Context, name string, workspaceID uuid.UUID) (*entity.Team, error) {
	team := &entity.Team{
		ID:          uuid.New(),
		Name:        name,
		WorkspaceID: workspaceID,
	}
	f.teams = append(f.teams, team)
	return team, nil
}

func (f *fakeTeamService) GetByID(_ context.Context, id uuid.UUID) (*entity.Team, error) {
	for _, t := range f.teams {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeTeamService) AddMember(_ context.Context, teamID, userID uuid.UUID, role string) error {
	for _, t := range f.teams {
		if t.ID == teamID {
			member := &entity.TeamMember{
				ID:       uuid.New(),
				TeamID:   teamID,
				UserID:   userID,
				Role:     role,
				JoinedAt: time.Now(),
			}
			f.members = append(f.members, member)
			return nil
		}
	}
	return errors.New("team not found")
}

func (f *fakeTeamService) RemoveMember(_ context.Context, teamID, userID uuid.UUID) error {
	for i, m := range f.members {
		if m.TeamID == teamID && m.UserID == userID {
			f.members = append(f.members[:i], f.members[i+1:]...)
			return nil
		}
	}
	return errors.New("member not found")
}

func (f *fakeTeamService) ListMembers(_ context.Context, teamID uuid.UUID) ([]*entity.TeamMember, error) {
	var result []*entity.TeamMember
	for _, m := range f.members {
		if m.TeamID == teamID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (f *fakeTeamService) AssignToWorkspace(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ entity.TeamWorkspaceRole) error {
	return nil
}

var _ service.TeamService = (*fakeTeamService)(nil)

func newTestTeamHandler(teams []*entity.Team, members []*entity.TeamMember) *TeamHandler {
	gin.SetMode(gin.TestMode)
	fs := &fakeTeamService{teams: teams, members: members}
	return NewTeamHandler(fs)
}

func newTestTeam() *entity.Team {
	return &entity.Team{
		ID:          uuid.New(),
		Name:        "Test Team",
		WorkspaceID: uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func performTeamRequest(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var payload *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		payload = bytes.NewReader(b)
	} else {
		payload = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, payload)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCreateTeamSuccessfully(t *testing.T) {
	h := newTestTeamHandler(nil, nil)
	r := gin.New()
	r.POST("/workspaces/:ws_id/teams", func(c *gin.Context) { h.Create(c) })

	reqBody := map[string]any{"name": "My Team"}
	rec := performTeamRequest(r, http.MethodPost, "/workspaces/"+uuid.New().String()+"/teams", reqBody)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestCreateTeamReturns400OnMissingName(t *testing.T) {
	h := newTestTeamHandler(nil, nil)
	r := gin.New()
	r.POST("/workspaces/:ws_id/teams", func(c *gin.Context) { h.Create(c) })

	rec := performTeamRequest(r, http.MethodPost, "/workspaces/"+uuid.New().String()+"/teams", map[string]any{"description": "No name"})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetByIDReturnsTeam(t *testing.T) {
	team := newTestTeam()
	h := newTestTeamHandler([]*entity.Team{team}, nil)
	r := gin.New()
	r.GET("/teams/:id", func(c *gin.Context) { h.GetByID(c) })

	rec := performTeamRequest(r, http.MethodGet, "/teams/"+team.ID.String(), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestTeamGetByIDReturns404WhenNotFound(t *testing.T) {
	h := newTestTeamHandler(nil, nil)
	r := gin.New()
	r.GET("/teams/:id", func(c *gin.Context) { h.GetByID(c) })

	rec := performTeamRequest(r, http.MethodGet, "/teams/"+uuid.New().String(), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAddMemberSuccessfully(t *testing.T) {
	team := newTestTeam()
	h := newTestTeamHandler([]*entity.Team{team}, nil)
	r := gin.New()
	r.POST("/teams/:id/members", func(c *gin.Context) { h.AddMember(c) })

	reqBody := map[string]any{"user_id": uuid.New().String(), "role": "member"}
	rec := performTeamRequest(r, http.MethodPost, "/teams/"+team.ID.String()+"/members", reqBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestListMembersReturnsMembers(t *testing.T) {
	team := newTestTeam()
	member := &entity.TeamMember{
		ID:       uuid.New(),
		TeamID:   team.ID,
		UserID:   uuid.New(),
		Role:     "member",
		JoinedAt: time.Now(),
	}
	h := newTestTeamHandler([]*entity.Team{team}, []*entity.TeamMember{member})
	r := gin.New()
	r.GET("/teams/:id/members", func(c *gin.Context) { h.ListMembers(c) })

	rec := performTeamRequest(r, http.MethodGet, "/teams/"+team.ID.String()+"/members", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestRemoveMemberSuccessfully(t *testing.T) {
	team := newTestTeam()
	member := &entity.TeamMember{
		ID:       uuid.New(),
		TeamID:   team.ID,
		UserID:   uuid.New(),
		Role:     "member",
		JoinedAt: time.Now(),
	}
	h := newTestTeamHandler([]*entity.Team{team}, []*entity.TeamMember{member})
	r := gin.New()
	r.DELETE("/teams/:id/members/:user_id", func(c *gin.Context) { h.RemoveMember(c) })

	rec := performTeamRequest(r, http.MethodDelete, "/teams/"+team.ID.String()+"/members/"+member.UserID.String(), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}
