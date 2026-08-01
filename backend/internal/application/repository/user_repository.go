package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// UserRepositoryImpl implements the repository.UserRepository interface.
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db repository.DB, gormDB *gorm.DB) repository.UserRepository {
	return &UserRepositoryImpl{db: gormDB}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Roles").
		First(&user, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("find user by ID %s: %w", id, err)
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Roles").
		First(&user, "email = ?", email).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("find user by email %s: %w", email, err)
	}
	return &user, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Model(user).Updates(map[string]any{
		"full_name": user.FullName,
		"is_active": user.IsActive,
	}).Error; err != nil {
		return fmt.Errorf("update user %s: %w", user.ID, err)
	}
	return nil
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("delete user %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *UserRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*entity.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var users []*entity.User
	var total int64

	if err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Preload("Roles").
		Offset(offset).Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	return users, int(total), nil
}

func (r *UserRepositoryImpl) CountActive(ctx context.Context) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("is_active = ?", true).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count active users: %w", err)
	}
	return int(count), nil
}
