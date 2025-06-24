package store

import (
	"context"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"demo520/pkg/auth"
	"errors"
	"time"

	"github.com/asaskevich/govalidator"
	"gorm.io/gorm"
)

type UserStore interface {
	Create(ctx context.Context, user *model.UserM) error
	Update(ctx context.Context, user *model.UserM) error
	Delete(ctx context.Context, userUUID string) error
	Get(ctx context.Context, email string) (*model.UserM, error)
	List(ctx context.Context, offset int, limit int) (*[]model.UserM, error)
	ChangePassword(ctx context.Context, email string, oldPassword string, newPassword string) error
}

type userStore struct {
	db *gorm.DB
}

var _ UserStore = (*userStore)(nil)

func newUserStore(db *gorm.DB) *userStore {
	return &userStore{
		db: db,
	}
}

func (u *userStore) Create(ctx context.Context, user *model.UserM) error {
	log.C(ctx).Infow("Database operation: Create user",
		"email", user.Email,
		"user_uuid", user.UserUUID,
		"operation", "INSERT",
		"table", "users",
	)

	if user.Email == "" {
		log.C(ctx).Errorw("Database operation failed: email cannot be empty",
			"operation", "INSERT",
			"table", "users",
		)
		return errno.ErrInvalidParameter.SetMessage("email cannot be empty")
	}
	if !govalidator.IsEmail(user.Email) {
		log.C(ctx).Errorw("Database operation failed: invalid email format",
			"email", user.Email,
			"operation", "INSERT",
			"table", "users",
		)
		return errno.ErrInvalidParameter.SetMessage("invalid email format: %s", user.Email)
	}
	if user.Password == "" {
		log.C(ctx).Errorw("Database operation failed: password cannot be empty",
			"email", user.Email,
			"operation", "INSERT",
			"table", "users",
		)
		return errno.ErrInvalidParameter.SetMessage("password cannot be empty")
	}

	log.C(ctx).Infow("Encrypting user password",
		"email", user.Email,
		"user_uuid", user.UserUUID,
	)

	start := time.Now()
	err := u.db.Create(user).Error
	duration := time.Since(start)

	if err != nil {
		log.C(ctx).Errorw("Database operation failed",
			"email", user.Email,
			"user_uuid", user.UserUUID,
			"error", err.Error(),
			"operation", "INSERT",
			"table", "users",
			"duration_ms", duration.Milliseconds(),
		)
		return err
	}

	log.C(ctx).Infow("Database operation successful",
		"email", user.Email,
		"user_uuid", user.UserUUID,
		"operation", "INSERT",
		"table", "users",
		"duration_ms", duration.Milliseconds(),
	)

	return nil
}

func (u *userStore) Update(ctx context.Context, user *model.UserM) error {
	if user == nil {
		log.Errorw("user cannot be nil")
		return errno.ErrInvalidParameter.SetMessage("user cannot be nil")
	}
	if user.UserUUID == "" {
		log.Errorw("userUUID cannot be empty")
		return errno.ErrInvalidParameter.SetMessage("userUUID cannot be empty")
	}
	if !govalidator.IsUUIDv4(user.UserUUID) {
		log.Errorw("invalid UUIDv4 format", "userUUID", user.UserUUID)
		return errno.ErrInvalidParameter.SetMessage("invalid UUIDv4 format: %s", user.UserUUID)
	}
	return u.db.Model(&model.UserM{}).Where("userUUID = ?", user.UserUUID).Omit("userUUID").Updates(user).Error
}

func (u *userStore) ChangePassword(ctx context.Context, userUUID, oldPassword, newPassword string) error {
	log.C(ctx).Infow("Database operation: Change user password",
		"user_uuid", userUUID,
		"operation", "UPDATE",
		"table", "users",
		"field", "password",
	)

	if userUUID == "" {
		log.C(ctx).Errorw("Database operation failed: userUUID cannot be empty",
			"operation", "UPDATE",
			"table", "users",
		)
		return errno.ErrInvalidParameter.SetMessage("userUUID cannot be empty")
	}
	if oldPassword == "" {
		log.C(ctx).Errorw("Database operation failed: oldPassword cannot be empty",
			"user_uuid", userUUID,
			"operation", "UPDATE",
			"table", "users",
		)
		return errno.ErrInvalidParameter.SetMessage("oldPassword cannot be empty")
	}
	if newPassword == "" {
		log.C(ctx).Errorw("Database operation failed: newPassword cannot be empty",
			"user_uuid", userUUID,
			"operation", "UPDATE",
			"table", "users",
		)
		return errno.ErrInvalidParameter.SetMessage("newPassword cannot be empty")
	}

	log.C(ctx).Infow("Starting database transaction for password change",
		"user_uuid", userUUID,
	)

	start := time.Now()
	err := u.db.Transaction(func(tx *gorm.DB) error {
		log.C(ctx).Debugw("Fetching user for password verification",
			"user_uuid", userUUID,
		)

		var user model.UserM
		if err := tx.Where("user_uuid = ?", userUUID).First(&user).Error; err != nil {
			log.C(ctx).Errorw("Failed to get user in transaction",
				"user_uuid", userUUID,
				"error", err.Error(),
				"operation", "SELECT",
				"table", "users",
			)
			return err
		}

		log.C(ctx).Debugw("Verifying old password",
			"user_uuid", userUUID,
			"email", user.Email,
		)

		if !auth.VerifyPassword(oldPassword, user.Password) {
			log.C(ctx).Warnw("Old password verification failed",
				"user_uuid", userUUID,
				"email", user.Email,
			)
			return errno.ErrPasswordIncorrect
		}

		log.C(ctx).Infow("Encrypting new password",
			"user_uuid", userUUID,
			"email", user.Email,
		)

		log.C(ctx).Debugw("Updating password in database",
			"user_uuid", userUUID,
			"email", user.Email,
		)
		if err := tx.Model(&user).Update("password", newPassword).Error; err != nil {
			log.C(ctx).Errorw("Failed to update password in transaction",
				"user_uuid", userUUID,
				"email", user.Email,
				"error", err.Error(),
				"operation", "UPDATE",
				"table", "users",
			)
			return err
		}

		log.C(ctx).Infow("Password updated successfully in transaction",
			"user_uuid", userUUID,
			"email", user.Email,
		)
		return nil
	})

	duration := time.Since(start)
	if err != nil {
		log.C(ctx).Errorw("Database transaction failed",
			"user_uuid", userUUID,
			"error", err.Error(),
			"operation", "UPDATE",
			"table", "users",
			"duration_ms", duration.Milliseconds(),
		)
		return err
	}

	log.C(ctx).Infow("Database transaction completed successfully",
		"user_uuid", userUUID,
		"operation", "UPDATE",
		"table", "users",
		"duration_ms", duration.Milliseconds(),
	)

	return nil
}

func (u *userStore) Delete(ctx context.Context, userUUID string) error {
	if userUUID == "" {
		log.Errorw("userUUID cannot be empty")
		return errno.ErrInvalidParameter.SetMessage("userUUID cannot be empty")
	}
	if !govalidator.IsUUIDv4(userUUID) {
		log.Errorw("invalid UUIDv4 format", "userUUID", userUUID)
		return errno.ErrInvalidParameter.SetMessage("invalid UUIDv4 format: %s", userUUID)
	}
	err := u.db.Model(&model.UserM{}).Where("userUUID = ?", userUUID).Delete(&model.UserM{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrUserNotFound.SetMessage("user not found with UUID: %s", userUUID)
		}
		return errno.InternalServerError.SetMessage("database error: %v", err)
	}
	return nil
}

func (u *userStore) Get(ctx context.Context, email string) (*model.UserM, error) {
	log.C(ctx).Infow("Database query: Get user by email",
		"email", email,
		"operation", "SELECT",
		"table", "users",
	)

	if email == "" {
		log.C(ctx).Errorw("Database query failed: email is empty",
			"operation", "SELECT",
			"table", "users",
		)
		return nil, errno.ErrInvalidParameter
	}

	var user model.UserM
	start := time.Now()
	if err := u.db.Where("email = ?", email).First(&user).Error; err != nil {
		duration := time.Since(start)
		log.C(ctx).Errorw("Database query failed",
			"email", email,
			"error", err.Error(),
			"operation", "SELECT",
			"table", "users",
			"duration_ms", duration.Milliseconds(),
			"sql_error", errors.Is(err, gorm.ErrRecordNotFound),
		)
		return nil, err
	}

	duration := time.Since(start)
	log.C(ctx).Infow("Database query successful",
		"email", email,
		"user_uuid", user.UserUUID,
		"operation", "SELECT",
		"table", "users",
		"duration_ms", duration.Milliseconds(),
		"created_at", user.CreatedAt,
	)

	return &user, nil
}

func (u *userStore) List(ctx context.Context, offset int, limit int) (*[]model.UserM, error) {
	if offset < 0 {
		log.Errorw("offset cannot be negative")
		return nil, errno.ErrInvalidParameter.SetMessage("offset cannot be negative: %d", offset)
	}
	if limit <= 0 {
		log.Errorw("limit must be positive")
		return nil, errno.ErrInvalidParameter.SetMessage("limit must be positive: %d", limit)
	}
	var users []model.UserM
	err := u.db.Model(&model.UserM{}).Limit(limit).Offset(offset).Find(&users).Error
	return &users, err
}
