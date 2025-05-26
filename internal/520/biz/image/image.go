package image

import (
	"context"
	"demo520/pkg/api"
)

type ImageBiz interface {
	Create(ctx context.Context, userUUID string, r *api.CreateImageRequest) (*api.CreateImageResponse, error)
	UpdateTags(ctx context.Context, userUUID string, imageUUID string, r *api.UpdateImageTagsRequest) error
	Delete(ctx context.Context, userUUID string, imageUUID string) error
	DeleteCollection(ctx context.Context, userUUID string, imageUUIDs []string) error
	Get(ctx context.Context, userUUID string, imageUUID string) (*api.GetImageInfoResponse, error)
	ListUserOwnImages(ctx context.Context, userUUID string, offset, limit int) (*api.ListImageResponse, error)
	ListUserOwnPublicImages(ctx context.Context, offset, limit int) (*api.ListImageResponse, error)
	ListRandomPublicImages(ctx context.Context, limit int) (*api.ListImageResponse, error)
}
