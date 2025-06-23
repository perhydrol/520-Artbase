package user

import (
	"context"
	"demo520/internal/520/store"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"demo520/pkg/api"
	"demo520/pkg/auth"
	"demo520/pkg/token"
	"errors"
	"regexp"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserBiz interface {
	ChangePassword(ctx context.Context, email string, r *api.ChangePasswordRequest) error
	Login(ctx context.Context, r *api.LoginRequest) (*api.LoginResponse, error)
	Create(ctx context.Context, r *api.CreateUserRequest) error
	Get(ctx context.Context, email string) (*api.GetUserInfoResponse, error)
	Update(ctx context.Context, userUUID, email string, r *api.UpdateUserRequest) error
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
	defer log.FuncEntryWithContext(ctx, email, "***")()

	if err := u.db.User().ChangePassword(ctx, email, r.OldPassword, r.NewPassword); err != nil {
		// ChangePassword方法内部已经处理了密码错误的情况，这里只处理其他错误
		if errors.Is(err, errno.ErrPasswordIncorrect) {
			log.ErrorWithFunc(err, "旧密码验证失败", "email", email)
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := errno.ErrUserNotFound
			log.ErrorWithFunc(err, "用户不存在", "email", email)
			return err
		}
		log.ErrorWithFunc(err, "修改密码失败", "email", email)
		return errno.InternalServerError.SetMessage("failed to change password: %v", err)
	}
	log.Infow("密码修改成功", "email", email)
	return nil
}

func (u *userBiz) Login(ctx context.Context, r *api.LoginRequest) (*api.LoginResponse, error) {
	defer log.FuncEntryWithContext(ctx, r.Email, "***", r.SeedTime)()

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
		err := errno.ErrUserLoginRequestOutTime
		log.ErrorWithFunc(err, "登录请求时间超时",
			"email", r.Email,
			"seedTime", r.SeedTime,
			"duration", duration.String())
		return nil, err
	}

	userM, err := u.db.User().Get(ctx, r.Email)
	if err != nil {
		log.ErrorWithFunc(err, "获取用户信息失败", "email", r.Email)
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
	defer log.FuncEntryWithContext(ctx, r.Email, "***")()

	var userM model.UserM
	userM.Email = r.Email
	userM.Nickname = r.Nickname
	userM.Password = r.Password
	userM.UserUUID = uuid.New().String()

	if err := u.db.User().Create(ctx, &userM); err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key 'username'", err.Error()); match {
			return errno.ErrUserAlreadyExist
		}
		return errno.InternalServerError.SetMessage("failed to create user: %v", err)
	}
	return nil
}

func (u *userBiz) Get(ctx context.Context, email string) (*api.GetUserInfoResponse, error) {
	defer log.FuncEntryWithContext(ctx, email)()

	user, err := u.db.User().Get(ctx, email)
	if err != nil {
		log.ErrorWithFunc(err, "获取用户信息失败", "email", email)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrUserNotFound
		}
		return nil, errno.InternalServerError.SetMessage("failed to get user: %v", err)
	}
	var resp api.GetUserInfoResponse
	resp.UserUUID = user.UserUUID
	resp.Email = user.Email
	resp.Nickname = user.Nickname
	resp.CreateAt = user.CreatedAt.Format(time.RFC3339)

	return &resp, nil
}

func (u *userBiz) Update(ctx context.Context, userUUID, email string, r *api.UpdateUserRequest) error {
	if userUUID == "" || email == "" {
		return errno.ErrInvalidParameter.SetMessage("missing required parameters: userUUID=%s, email=%s", userUUID, email)
	}
	if !govalidator.IsEmail(email) {
		return errno.ErrInvalidParameter.SetMessage("invalid email format: %s", email)
	}
	if !govalidator.IsUUIDv4(userUUID) {
		return errno.ErrInvalidParameter.SetMessage("invalid uuid format: %s", userUUID)
	}
	userM, err := u.db.User().Get(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrUserNotFound
		}
		return errno.InternalServerError.SetMessage("failed to get user: %v", err)
	}
	if userUUID != userM.UserUUID {
		return errno.ErrUnauthorized.SetMessage("operation not permitted: userUUID mismatch (got %s)", userUUID)
	}

	if r.Email != "" {
		userM.Email = r.Email
	}

	if r.Nickname != "" {
		userM.Nickname = r.Nickname
	}

	if err := u.db.User().Update(ctx, userM); err != nil {
		return errno.InternalServerError.SetMessage("failed to update user: %v", err)
	}

	return nil
}

func (u *userBiz) Delete(ctx context.Context, userUUID string) error {
	if err := u.db.User().Delete(ctx, userUUID); err != nil {
		return errno.InternalServerError.SetMessage("failed to delete user: %v", err)
	}
	return nil
}
