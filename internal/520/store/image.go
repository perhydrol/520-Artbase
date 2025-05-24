package store

import (
	"context"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"errors"
	"gorm.io/gorm"
)

type ImageStore interface {
	Create(ctx *context.Context, image *model.ImageM) error
	Get(ctx *context.Context, imageUUID string) (*model.ImageM, error)
	Delete(ctx *context.Context, imageUUID string) error
	AddTagsToImage(ctx *context.Context, imageUUID string, tags []model.ImageTagM) error
	DeleteTagFromImage(ctx *context.Context, imageUUID string, tag model.ImageTagM) error
}

type imageStore struct {
	db *gorm.DB
}

var _ ImageStore = (*imageStore)(nil)

func newImageStore(db *gorm.DB) ImageStore {
	return &imageStore{
		db: db,
	}
}

func (u *imageStore) Create(ctx *context.Context, image *model.ImageM) error {
	return u.db.Create(image).Error
}

func (u *imageStore) Get(ctx *context.Context, imageUUID string) (*model.ImageM, error) {
	var image model.ImageM
	err := u.db.First(&image, "imageUUID = ?", imageUUID).Error
	return &image, err
}

func (u *imageStore) Delete(ctx *context.Context, imageUUID string) error {
	err := u.db.Delete(&model.ImageM{}, "imageUUID = ?", imageUUID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (u *imageStore) AddTagsToImage(ctx *context.Context, imageUUID string, tags []model.ImageTagM) error {
	image, err := u.Get(ctx, imageUUID)
	if err != nil {
		return err
	} else if len(tags) < 0 {
		log.Warnw("tag length is zero",
			"imageUUID", imageUUID)
		return errors.New("tag length is zero")
	}
	err = u.db.Model(image).Association("tags").Append(tags)
	if err != nil {
		return err
	}
	return nil
}

func (u *imageStore) DeleteTagFromImage(ctx *context.Context, imageUUID string, tag model.ImageTagM) error {
	image, err := u.Get(ctx, imageUUID)
	if err != nil {
		return err
	}
	err = u.db.Model(image).Association("tags").Delete(tag)
	if err != nil {
		return err
	}
	return nil
}
