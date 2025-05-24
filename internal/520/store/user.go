package store

import (
	"context"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"errors"
	"gorm.io/gorm"
)

type UserStore interface {
	Create(ctx context.Context, user *model.UserM) error
	Update(ctx context.Context, user *model.UserM) error
	Delete(ctx context.Context, userUUID string) error
	Get(ctx context.Context, userName string) (*model.UserM, error)
	List(ctx context.Context, offset int, limit int) (*[]model.UserM, error)
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
		log.Errorw("UserStore is nil")
	}
	return u.db.Create(user).Error
}

func (u *userStore) Update(ctx context.Context, user *model.UserM) error {
	if _, err := u.Get(ctx, user.UserUUID); err != nil {
		return err
	}
	return u.db.Save(user).Error
}

func (u *userStore) Delete(ctx context.Context, userUUID string) error {
	if _, err := u.Get(ctx, userUUID); err != nil {
		return nil
	}
	err := u.db.Delete(&model.UserM{UserUUID: userUUID}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func (u *userStore) Get(ctx context.Context, userUUID string) (*model.UserM, error) {
	var user model.UserM
	err := u.db.First(&user, "userUUID = ?", userUUID).Error
	return &user, err
}

func (u *userStore) List(ctx context.Context, offset int, limit int) (*[]model.UserM, error) {
	err := u.db.Limit(limit).Offset(offset).Find(&[]model.UserM{}).Error
	return &[]model.UserM{}, err
}
