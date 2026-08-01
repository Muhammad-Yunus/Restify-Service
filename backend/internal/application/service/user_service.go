package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// UserService implements the service.UserService interface.
type UserService struct {
	db     *gorm.DB
	logger repository.Logger
}

// NewUserService creates a new user service.
func NewUserService(gormDB *gorm.DB, logger repository.Logger) *UserService {
	return &UserService{
		db:     gormDB,
		logger: logger,
	}
}

// repository returns a new UserRepositoryImpl.
func (s *UserService) repository() repository.UserRepository {
	return &userRepositoryImpl{db: s.db}
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	repo := s.repository()
	user, err := repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}
	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	repo := s.repository()
	user, err := repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email %s: %w", email, err)
	}
	return user, nil
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.User, error) {
	repo := s.repository()
	user, err := repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name, ok := updates["full_name"].(string); ok {
		user.FullName = &name
	}
	if active, ok := updates["is_active"].(bool); ok {
		user.IsActive = active
	}

	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("validate user: %w", err)
	}

	if err := repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user %s: %w", id, err)
	}

	s.logger.Info(ctx, "user updated", "user_id", id)
	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	repo := s.repository()
	if err := repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user %s: %w", id, err)
	}
	s.logger.Info(ctx, "user deleted", "user_id", id)
	return nil
}

func (s *UserService) List(ctx context.Context, page, pageSize int) ([]*entity.User, int, error) {
	repo := s.repository()
	users, total, err := repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}
