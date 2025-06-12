package store

import (
	"context"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"demo520/pkg/auth"
	"errors"
	"fmt"

	"github.com/asaskevich/govalidator"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	if user == nil {
		log.Errorw("user cannot be nil")
		return errors.New("user cannot be nil")
	}
	return u.db.Create(user).Error
}

func (u *userStore) Update(ctx context.Context, user *model.UserM) error {
	if user == nil {
		log.Errorw("user cannot be nil")
		return errors.New("user cannot be nil")
	}
	if user.UserUUID == "" {
		log.Errorw("userUUID cannot be empty")
		return errors.New("userUUID cannot be empty")
	}
	if !govalidator.IsUUIDv4(user.UserUUID) {
		log.Errorw("invalid UUIDv4 format", "userUUID", user.UserUUID)
		return errors.New("invalid UUIDv4 format")
	}
	return u.db.Model(&model.UserM{}).Where("userUUID = ?", user.UserUUID).Omit("userUUID").Updates(user).Error
}

func (u *userStore) ChangePassword(ctx context.Context, email string, oldPassword string, newPassword string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		var user model.UserM
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("email = ?", email).
			First(&user).Error; err != nil {
			return err
		}

		if !auth.VerifyPassword(oldPassword, user.Password) {
			return errno.ErrPasswordIncorrect
		}
		newHash, err := auth.HashPassword(newPassword)
		if err != nil {
			return err
		}
		if err = tx.Model(&user).Where("email = ?", email).Update("password", newHash).Error; err != nil {
			return err
		}
		return nil
	})
}

func (u *userStore) Delete(ctx context.Context, userUUID string) error {
	if userUUID == "" {
		log.Errorw("userUUID cannot be empty")
		return errors.New("userUUID cannot be empty")
	}
	if !govalidator.IsUUIDv4(userUUID) {
		log.Errorw("invalid UUIDv4 format", "userUUID", userUUID)
		return errors.New("invalid UUIDv4 format")
	}
	err := u.db.Model(&model.UserM{}).Where("userUUID = ?", userUUID).Delete(&model.UserM{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found with UUID: %s", userUUID)
		}
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

func (u *userStore) Get(ctx context.Context, email string) (*model.UserM, error) {
	if email == "" {
		log.Errorw("email cannot be empty")
		return nil, errors.New("email cannot be empty")
	}
	if !govalidator.IsEmail(email) {
		log.Errorw("invalid email format", "email", email)
		return nil, errors.New("invalid email format")
	}
	var user model.UserM
	err := u.db.Model(&model.UserM{}).First(&user, "email = ?", email).Error
	return &user, err
}

func (u *userStore) List(ctx context.Context, offset int, limit int) (*[]model.UserM, error) {
	if offset < 0 {
		log.Errorw("offset cannot be negative")
		return nil, errors.New("offset cannot be negative")
	}
	if limit <= 0 {
		log.Errorw("limit must be positive")
		return nil, errors.New("limit must be positive")
	}
	var users []model.UserM
	err := u.db.Model(&model.UserM{}).Limit(limit).Offset(offset).Find(&users).Error
	return &users, err
}
