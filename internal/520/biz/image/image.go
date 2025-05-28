package image

import (
	"context"
	"demo520/internal/520/store"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/model"
	"demo520/pkg/api"
	"fmt"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"mime/multipart"
)

type ImageBiz interface {
	Create(ctx *context.Context, userUUID string, r *api.CreateImageRequest, fileHeader *multipart.FileHeader) (*api.CreateImageResponse, error)
	UpdateTags(ctx *context.Context, userUUID string, imageUUID string, r *api.UpdateImageTagsRequest) error
	Delete(ctx *context.Context, userUUID string, imageUUID string) error
	DeleteCollection(ctx *context.Context, userUUID string, imageUUIDs []string) error
	Get(ctx *context.Context, userUUID string, imageUUID string) (*api.GetImageInfoResponse, error)
	ListUserOwnImages(ctx *context.Context, userUUID string, offset, limit int) (*api.ListImageResponse, error)
	ListUserOwnPublicImages(ctx *context.Context, offset, limit int) (*api.ListImageResponse, error)
	ListRandomPublicImages(ctx *context.Context, limit int) (*api.ListImageResponse, error)
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

func (i *imageBiz) Create(ctx *context.Context, userUUID string, r *api.CreateImageRequest, fileHeader *multipart.FileHeader) (*api.CreateImageResponse, error) {
	if fileHeader == nil || r == nil {
		return nil, fmt.Errorf("file or request is nil")
	}

	if ok, err := i.imageFileStore.Validate(fileHeader); err != nil {
		return nil, fmt.Errorf("failed to validate image: %w", err)
	} else if !ok {
		return nil, errno.ErrImageFileInvalid
	}

	imageMaxSize := viper.GetInt64("ImageMaxSize")
	if fileHeader.Size > imageMaxSize {
		return nil, errno.ErrImageFileTooLarge
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
		OwnerUUID: r.OwnerUUID,
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

func (i *imageBiz) UpdateTags(ctx *context.Context, userUUID string, imageUUID string, r *api.UpdateImageTagsRequest) error {
	//TODO implement me
	panic("implement me")
}

func (i *imageBiz) Delete(ctx *context.Context, userUUID string, imageUUID string) error {
	//TODO implement me
	panic("implement me")
}

func (i *imageBiz) DeleteCollection(ctx *context.Context, userUUID string, imageUUIDs []string) error {
	//TODO implement me
	panic("implement me")
}

func (i *imageBiz) Get(ctx *context.Context, userUUID string, imageUUID string) (*api.GetImageInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i *imageBiz) ListUserOwnImages(ctx *context.Context, userUUID string, offset, limit int) (*api.ListImageResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i *imageBiz) ListUserOwnPublicImages(ctx *context.Context, offset, limit int) (*api.ListImageResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i *imageBiz) ListRandomPublicImages(ctx *context.Context, limit int) (*api.ListImageResponse, error) {
	//TODO implement me
	panic("implement me")
}
