package store

import (
	"context"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/model"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math/rand"
)

type ImageStore interface {
	Create(ctx context.Context, image *model.ImageM) error
	Get(ctx context.Context, imageUUID string) (*model.ImageM, error)
	Delete(ctx context.Context, imageUUID string) error
	AddTagsToImage(ctx context.Context, imageUUID string, tags []string) error
	DeleteTagFromImage(ctx context.Context, imageUUID string, tag []string) error
	GetUserImages(ctx context.Context, UserUUID string, offset, limit int) (int64, []*model.ImageM, error)
	GetRandomPublicImages(ctx context.Context, limit int) (int, []*model.ImageM, error)
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

func (u *imageStore) Create(ctx context.Context, image *model.ImageM) error {
	return u.db.Create(image).Error
}

func (u *imageStore) Get(ctx context.Context, imageUUID string) (*model.ImageM, error) {
	var image model.ImageM
	err := u.db.Preload("Tags").First(&image, "imageUUID = ?", imageUUID).Error
	return &image, err
}

func (u *imageStore) Delete(ctx context.Context, imageUUID string) error {
	err := u.db.Delete(&model.ImageM{}, "imageUUID = ?", imageUUID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	u.db.Delete(&model.ImageTagM{}, "imageUUID = ?", imageUUID)
	return nil
}

func (u *imageStore) AddTagsToImage(ctx context.Context, imageUUID string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	err := u.db.Transaction(func(tx *gorm.DB) error {
		var image model.ImageM
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&image, "imageUUID = ?", imageUUID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: image not found", errno.ErrImageNotFound)
			}
			return fmt.Errorf("failed to lock image: %w", err)
		}
		var existingTags []model.ImageTagM
		if err := tx.Where("imageUUID = ?", imageUUID).Find(&existingTags).Error; err != nil {
			return fmt.Errorf("failed to find existing tags: %w", err)
		}
		existingTagMap := make(map[string]struct{}, len(existingTags))
		for _, tag := range existingTags {
			existingTagMap[tag.Tag] = struct{}{}
		}
		uniqueTags := make([]model.ImageTagM, 0)
		for _, tag := range tags {
			if _, exists := existingTagMap[tag]; !exists {
				uniqueTags = append(uniqueTags, model.ImageTagM{Tag: tag})
			}
		}
		err := tx.Model(&image).Association("Tags").Append(uniqueTags)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *imageStore) DeleteTagFromImage(ctx context.Context, imageUUID string, tag []string) error {
	err := u.db.Transaction(func(tx *gorm.DB) error {
		var image model.ImageM
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("imageUUID=?", imageUUID).First(&image).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
			}
		}
		if err := tx.Where("imageUUID=? AND tag IN ?", imageUUID, tag).Delete(&model.ImageTagM{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *imageStore) GetRandomPublicImages(ctx context.Context, limit int) (retCount int, ret []*model.ImageM, err error) {
	var allCount int64
	if err := u.db.Model(&model.ImageM{}).Where("is_public = ?", true).Count(&allCount).Error; err != nil {
		return 0, nil, err
	}
	if allCount == 0 {
		return 0, nil, nil
	}
	var offset int
	if allCount <= int64(limit) {
		retCount = int(allCount)
		offset = 0
	} else {
		retCount = limit
		offset = rand.Intn(int(allCount) - retCount)
	}
	err = u.db.Model(&model.ImageM{}).Where("is_public = ?", true).Offset(offset).Limit(limit).Find(&ret).Error
	return
}

func (u *imageStore) GetUserImages(ctx context.Context, UserUUID string, offset, limit int) (count int64, ret []*model.ImageM, err error) {
	err = u.db.Model(&model.ImageM{}).Where("userUUID = ?", UserUUID).Offset(offset).Limit(limit).Find(&ret).Count(&count).Error
	return
}
