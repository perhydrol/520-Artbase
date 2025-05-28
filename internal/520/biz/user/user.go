package user

import (
	"context"
	"demo520/internal/520/store"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/model"
	"demo520/pkg/api"
	"demo520/pkg/auth"
	"demo520/pkg/token"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"regexp"
	"time"
)

type UserBiz interface {
	ChangePassword(ctx context.Context, email string, r *api.ChangePasswordRequest) error
	Login(ctx context.Context, r *api.LoginRequest) (*api.LoginResponse, error)
	Create(ctx context.Context, r *api.CreateUserRequest) error
	Get(ctx context.Context, userUUID string) (*api.GetUserInfoResponse, error)
	Update(ctx context.Context, userUUID string, r *api.UpdateUserRequest) error
	Delete(ctx context.Context, userUUID string) error
}

var _ UserBiz = (*userBiz)(nil)

type userBiz struct {
	db store.IStore
}

func NewUserBiz(db store.IStore) UserBiz {
	return &userBiz{
		db: db,
	}
}

func (u *userBiz) ChangePassword(ctx context.Context, email string, r *api.ChangePasswordRequest) error {
	userM, err := u.db.User().Get(ctx, email)
	if err != nil {
		return err
	}

	if !auth.VerifyPassword(r.OldPassword, userM.Password) {
		return errno.ErrPasswordIncorrect
	}
	userM.Password, err = auth.HashPassword(r.NewPassword)
	if err != nil {
		return err
	}
	if err = u.db.User().Update(ctx, userM); err != nil {
		return err
	}
	return nil
}

func (u *userBiz) Login(ctx context.Context, r *api.LoginRequest) (*api.LoginResponse, error) {
	var t time.Time

	// 自动检测时间戳单位（秒或毫秒）
	switch {
	case r.SeedTime > 1e18: // 纳秒（通常不需要处理）
		t = time.Unix(0, r.SeedTime)
	case r.SeedTime > 1e15: // 微秒
		t = time.Unix(0, r.SeedTime*1e3)
	case r.SeedTime > 1e12: // 毫秒
		t = time.Unix(r.SeedTime/1000, (r.SeedTime%1000)*1e6)
	default: // 秒
		t = time.Unix(r.SeedTime, 0)
	}

	now := time.Now().UTC()
	utcTime := t.UTC()
	duration := now.Sub(utcTime)
	if duration < 0 || duration >= 5*time.Minute {
		return nil, errno.ErrUserLoginRequestOutTime
	}

	userM, err := u.db.User().Get(ctx, r.Email)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	if !auth.VerifyPassword(r.Password, userM.Password) {
		return nil, errno.ErrPasswordIncorrect
	}

	jwt, err := token.GenerateToken(userM.UserUUID)
	if err != nil {
		return nil, err
	}
	return &api.LoginResponse{Token: jwt}, nil
}

func (u *userBiz) Create(ctx context.Context, r *api.CreateUserRequest) error {
	var userM model.UserM
	userM.Email = r.Email
	userM.Nickname = r.Nickname
	userM.Password = r.Password
	userM.UserUUID = uuid.New().String()

	if err := u.db.User().Create(ctx, &userM); err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key 'username'", err.Error()); match {
			return errno.ErrUserAlreadyExist
		}

		return err
	}
	return nil
}

func (u *userBiz) Get(ctx context.Context, userUUID string) (*api.GetUserInfoResponse, error) {
	user, err := u.db.User().Get(ctx, userUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrUserNotFound
		}
		return nil, err
	}
	var resp api.GetUserInfoResponse
	resp.UserUUID = user.UserUUID
	resp.Email = user.Email
	resp.Nickname = user.Nickname
	resp.CreateAt = user.CreatedAt.Format(time.RFC3339)

	return &resp, nil
}

func (u *userBiz) Update(ctx context.Context, userUUID string, r *api.UpdateUserRequest) error {
	userM, err := u.db.User().Get(ctx, userUUID)
	if err != nil {
		return err
	}

	if r.Email != "" {
		userM.Email = r.Email
	}

	if r.Nickname != "" {
		userM.Nickname = r.Nickname
	}

	if err := u.db.User().Update(ctx, userM); err != nil {
		return err
	}

	return nil
}

func (u *userBiz) Delete(ctx context.Context, userUUID string) error {
	if err := u.db.User().Delete(ctx, userUUID); err != nil {
		return err
	}
	return nil
}
