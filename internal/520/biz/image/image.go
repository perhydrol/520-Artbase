package image

import (
	"context"
	"demo520/internal/520/store"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/model"
	"demo520/pkg/api"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type ImageBiz interface {
	Create(ctx context.Context, userUUID string, r *api.CreateImageRequest, fileHeader *multipart.FileHeader) (*api.CreateImageResponse, error)
	UpdateTags(ctx context.Context, userUUID string, imageUUID string, r *api.UpdateImageTagsRequest) error
	Delete(ctx context.Context, userUUID string, imageUUID string) error
	DeleteCollection(ctx context.Context, userUUID string, imageUUIDs []string) error
	Get(ctx context.Context, userUUID string, imageUUID string) (*api.GetImageInfoResponse, error)
	ListUserOwnImages(ctx context.Context, userUUID string, offset, limit int) (*api.ListImageResponse, error)
	ListUserOwnPublicImages(ctx context.Context, userUUID string, offset, limit int) (*api.ListImageResponse, error)
	ListRandomPublicImages(ctx context.Context, limit int) (*api.ListImageResponse, error)
}

type imageBiz struct {
	db             store.IStore
	imageFileStore ImageFileStore
}

var _ ImageBiz = (*imageBiz)(nil)

func NewImageBiz(db store.IStore) ImageBiz {
	return &imageBiz{
		db:             db,
		imageFileStore: NewImageFileStore(),
	}
}

func (i *imageBiz) Create(ctx context.Context, userUUID string, r *api.CreateImageRequest, fileHeader *multipart.FileHeader) (*api.CreateImageResponse, error) {
	if fileHeader == nil {
		return nil, fmt.Errorf("%w: file header", errno.ErrInvalidParameter)
	}
	if r == nil {
		return nil, fmt.Errorf("%w: request", errno.ErrInvalidParameter)
	}

	imageMaxSize := viper.GetInt64("ImageMaxSize")
	if fileHeader.Size > imageMaxSize {
		return nil, errno.ErrImageFileTooLarge
	}

	if ok, err := i.imageFileStore.Validate(fileHeader); err != nil {
		return nil, fmt.Errorf("failed to validate image: %w", err)
	} else if !ok {
		return nil, errno.ErrImageFileInvalid
	}

	hash, err := i.imageFileStore.Hash(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate image hash: %w", err)
	}

	imageUUID := uuid.New().String()
	if err := i.imageFileStore.Save(fileHeader, hash); err != nil {
		return nil, fmt.Errorf("failed to save image file: %w", err)
	}

	imageM := model.ImageM{
		ImageUUID: imageUUID,
		Hash:      hash,
		Token:     "",
		UserUUID:  r.UserUUID,
		IsPublic:  r.IsPublic,
		Tags:      nil,
	}
	if err := i.db.Image().Create(ctx, &imageM); err != nil {
		return nil, fmt.Errorf("failed to create image record: %w", err)
	}

	var ret api.CreateImageResponse
	if err := copier.Copy(&ret, imageM); err != nil {
		return nil, fmt.Errorf("failed to copy image data: %w", err)
	}
	return &ret, nil
}

func (i *imageBiz) UpdateTags(ctx context.Context, userUUID string, imageUUID string, r *api.UpdateImageTagsRequest) error {
	if !govalidator.IsUUID(imageUUID) {
		return fmt.Errorf("%w: invalid image UUID", errno.ErrInvalidParameter)
	}
	if !govalidator.IsUUID(userUUID) {
		return fmt.Errorf("%w: invalid user UUID", errno.ErrInvalidParameter)
	}
	if len(r.Tags) == 0 {
		return fmt.Errorf("%w: empty tags", errno.ErrInvalidParameter)
	}
	imageM, getImageErr := i.db.Image().Get(ctx, imageUUID)
	if getImageErr != nil {
		return getImageErr
	}
	if imageM.UserUUID != userUUID {
		return fmt.Errorf("%w: unauthorized operation", errno.ErrUnauthorized)
	}
	tagSet := make(map[string]struct{})
	var uniqueTags []string
	for _, tag := range r.Tags {
		normalized := strings.TrimSpace(tag)
		if _, ok := tagSet[normalized]; !ok {
			tagSet[tag] = struct{}{}
			uniqueTags = append(uniqueTags, normalized)
		}
	}
	addTagsErr := i.db.Image().AddTagsToImage(ctx, imageUUID, uniqueTags)
	if addTagsErr != nil {
		return addTagsErr
	}
	return nil
}

func (i *imageBiz) Delete(ctx context.Context, userUUID string, imageUUID string) error {
	// 参数校验
	if !govalidator.IsUUID(imageUUID) {
		return fmt.Errorf("%w: invalid image UUID", errno.ErrInvalidParameter)
	}
	if !govalidator.IsUUID(userUUID) {
		return fmt.Errorf("%w: invalid user UUID", errno.ErrInvalidParameter)
	}
	imageM, getImageErr := i.db.Image().Get(ctx, imageUUID)
	if getImageErr != nil {
		return getImageErr
	}
	if imageM.UserUUID != userUUID {
		return fmt.Errorf("%w: unauthorized operation", errno.ErrUnauthorized)
	}
	delErr := i.db.Image().Delete(ctx, imageUUID)
	if delErr != nil {
		return delErr
	}
	return nil
}

func (i *imageBiz) DeleteCollection(ctx context.Context, userUUID string, imageUUIDs []string) error {
	//TODO implement me
	panic("implement me")
}

func (i *imageBiz) Get(ctx context.Context, userUUID string, imageUUID string) (*api.GetImageInfoResponse, error) {
	if !govalidator.IsUUID(imageUUID) {
		return nil, fmt.Errorf("%w: invalid image UUID", errno.ErrInvalidParameter)
	}
	if !govalidator.IsUUID(userUUID) {
		return nil, fmt.Errorf("%w: invalid user UUID", errno.ErrInvalidParameter)
	}
	imageM, getImageErr := i.db.Image().Get(ctx, imageUUID)
	if getImageErr != nil {
		if errors.Is(getImageErr, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: image=%s", errno.ErrImageNotFound, imageUUID)
		}
		return nil, getImageErr
	}
	if !imageM.IsPublic && imageM.UserUUID != userUUID {
		return nil, fmt.Errorf("%w: unauthorized operation", errno.ErrUnauthorized)
	}
	var ret api.GetImageInfoResponse
	if err := copier.Copy(&ret, imageM); err != nil {
		return nil, fmt.Errorf("failed to copy image data: %w", err)
	}
	return &ret, nil
}

func (i *imageBiz) ListUserOwnImages(ctx context.Context, userUUID string, offset, limit int) (*api.ListImageResponse, error) {
	if offset < 0 {
		return nil, fmt.Errorf("%w: invalid offset", errno.ErrInvalidParameter)
	}
	if limit < 0 {
		return nil, fmt.Errorf("%w: invalid limit", errno.ErrInvalidParameter)
	}
	count, imageList, getImageErr := i.db.Image().GetUserImages(ctx, userUUID, offset, limit)
	if getImageErr != nil {
		return nil, getImageErr
	}
	if count > 0 && imageList[0].UserUUID != userUUID {
		return nil, fmt.Errorf("%w: unauthorized operation", errno.ErrUnauthorized)
	}
	if count == 0 {
		return &api.ListImageResponse{}, nil
	}
	var ret api.ListImageResponse
	ret.Count = int(count)
	imageInfos := make([]api.ImageInfo, len(imageList))
	for i, image := range imageList {
		copyErr := copier.Copy(&imageInfos[i], image)
		if copyErr != nil {
			return nil, copyErr
		}
	}
	ret.ImageList = imageInfos
	return &ret, nil
}

func (i *imageBiz) ListUserOwnPublicImages(ctx context.Context, userUUID string, offset, limit int) (*api.ListImageResponse, error) {
	if offset < 0 {
		return nil, fmt.Errorf("%w: invalid offset", errno.ErrInvalidParameter)
	}
	if limit < 0 {
		return nil, fmt.Errorf("%w: invalid limit", errno.ErrInvalidParameter)
	}
	count, imageList, getImageErr := i.db.Image().GetUserImages(ctx, userUUID, offset, limit)
	if getImageErr != nil {
		return nil, getImageErr
	}
	var ret api.ListImageResponse
	ret.Count = int(count)
	imageInfos := make([]api.ImageInfo, len(imageList))
	for i, image := range imageList {
		if !image.IsPublic {
			continue
		}
		copyErr := copier.Copy(&imageInfos[i], image)
		if copyErr != nil {
			return nil, copyErr
		}
	}
	ret.ImageList = imageInfos
	return &ret, nil
}

func (i *imageBiz) ListRandomPublicImages(ctx context.Context, limit int) (*api.ListImageResponse, error) {
	if limit < 0 {
		return nil, fmt.Errorf("%w: invalid limit", errno.ErrInvalidParameter)
	}
	count, imageList, getImageErr := i.db.Image().GetRandomPublicImages(ctx, limit)
	if getImageErr != nil {
		return nil, getImageErr
	}
	var ret api.ListImageResponse
	ret.Count = int(count)
	imageInfos := make([]api.ImageInfo, len(imageList))
	for i, image := range imageList {
		copyErr := copier.Copy(&imageInfos[i], image)
		if copyErr != nil {
			return nil, copyErr
		}
	}
	ret.ImageList = imageInfos
	return &ret, nil
}
