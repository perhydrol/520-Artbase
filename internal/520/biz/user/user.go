package user

import (
	"context"
	"demo520/pkg/api"
)

type UserBiz interface {
	ChangePassword(ctx context.Context, userUUID string, r *api.ChangePasswordRequest) error
	Login(ctx context.Context, r *api.LoginRequest) (*api.LoginResponse, error)
	Create(ctx context.Context, r *api.CreateUserRequest) error
	Get(ctx context.Context, userUUID string) (*api.GetUserInfoResponse, error)
	Update(ctx context.Context, userUUID string, r *api.UpdateUserRequest) error
	Delete(ctx context.Context, userUUID string) error
}
