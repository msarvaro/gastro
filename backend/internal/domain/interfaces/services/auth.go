package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (*entity.User, string, error)
	ValidateToken(ctx context.Context, token string) (*entity.User, error)
	RefreshToken(ctx context.Context, token string) (string, error)
	ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, username string) error
}
