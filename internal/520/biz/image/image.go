package image

import (
	"context"
	"demo520/internal/520/store"
	"demo520/internal/pkg/config"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"demo520/pkg/api"
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"gorm.io/datatypes"
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
	GetImageFile(ctx context.Context, userUUID string, imageUUIDFileName string) (filePath string, err error)
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

func copyImageInfo(info *api.ImageInfo, imageM *model.NewImageM) error {
	info.ImageUUID = imageM.ImageUUID.String()
	if len(imageM.Token) != 0 {
		info.Token = string(imageM.Token)
	} else {
		info.Token = ""
	}
	info.UserUUID = imageM.UserUUID.String()
	info.IsPublic = imageM.IsPublic
	if len(imageM.Tags) > 0 {
		info.Tags = make([]string, len(imageM.Tags))
		for i, t := range imageM.Tags {
			info.Tags[i] = t.Tag
		}
	}
	if !imageM.CreatedAt.IsZero() {
		info.CreatedAt = imageM.CreatedAt.String()
	}
	if !imageM.UpdatedAt.IsZero() {
		info.UpdatedAt = imageM.UpdatedAt.String()
	}
	return nil
}

func (i *imageBiz) Create(ctx context.Context, userUUID string, r *api.CreateImageRequest, fileHeader *multipart.FileHeader) (*api.CreateImageResponse, error) {
	defer log.FuncEntryWithContext(ctx, userUUID, r, fileHeader)()

	// 参数验证
	if fileHeader == nil {
		err := errno.ErrInvalidParameter.SetMessage("file header is required")
		log.ErrorWithFunc(err, "参数验证失败", "parameter", "fileHeader")
		return nil, err
	}
	if r == nil {
		err := errno.ErrInvalidParameter.SetMessage("request is required")
		log.ErrorWithFunc(err, "参数验证失败", "parameter", "request")
		return nil, err
	}
	if len(r.Tags) > config.GetImage().MaxTagCount {
		err := errno.ErrInvalidParameter.SetMessage("Image's tags too much, max count is %d,action %d", config.GetImage().MaxTagCount, len(r.Tags))
		log.ErrorWithFunc(err, "创建图片时标签过多", "parameter", "request")
		return nil, err
	}

	// 文件大小验证
	imageMaxSize := config.GetImage().ImageMaxSize
	if fileHeader.Size > imageMaxSize {
		err := errno.ErrImageFileTooLarge
		log.ErrorWithFunc(err, "文件大小超限",
			"fileSize", fileHeader.Size,
			"maxSize", imageMaxSize)
		return nil, err
	}

	// 文件格式验证
	if ok, err := i.imageFileStore.Validate(fileHeader); err != nil {
		log.ErrorWithFunc(err, "文件格式验证失败")
		return nil, errno.InternalServerError.SetMessage("failed to validate image: %v", err)
	} else if !ok {
		err := errno.ErrImageFileInvalid
		log.ErrorWithFunc(err, "文件格式无效")
		return nil, err
	}

	// 计算文件哈希
	hash, err := i.imageFileStore.Hash(fileHeader)
	if err != nil {
		log.ErrorWithFunc(err, "计算文件哈希失败")
		return nil, errno.InternalServerError.SetMessage("failed to calculate image hash: %v", err)
	}
	log.Infow("文件哈希计算成功", "hash", hash)

	// 保存文件
	imageUUID := uuid.New()
	if err := i.imageFileStore.Save(fileHeader, hash); err != nil {
		log.ErrorWithFunc(err, "保存文件失败", "imageUUID", imageUUID.String())
		return nil, errno.InternalServerError.SetMessage("failed to save image file: %v", err)
	}
	log.Infow("文件保存成功", "imageUUID", imageUUID.String(), "hash", hash)

	// 处理标签
	var imageTags []model.ImageTagM
	for _, tag := range r.Tags {
		imageTags = append(imageTags, model.ImageTagM{
			Tag:       tag,
			ImageUUID: datatypes.BinUUID(imageUUID),
		})
	}

	// 解析用户UUID
	userUUIDBin, err := uuid.Parse(userUUID)
	if err != nil {
		return nil, errno.ErrInvalidParameter.SetMessage("invalid user UUID: %v", err)
	}

	// 创建图片记录
	imageM := model.NewImageM{
		ImageUUID: datatypes.BinUUID(imageUUID),
		Hash:      []byte(hash),
		Token:     []byte(""),
		UserUUID:  datatypes.BinUUID(userUUIDBin),
		IsPublic:  r.IsPublic,
		Tags:      imageTags,
	}
	if err := i.db.Image().Create(ctx, &imageM); err != nil {
		return nil, errno.InternalServerError.SetMessage("failed to create image record: %v", err)
	}

	// 复制响应数据
	var ret api.CreateImageResponse
	if err := copyImageInfo((*api.ImageInfo)(&ret), &imageM); err != nil {
		return nil, errno.InternalServerError.SetMessage("failed to copy image data: %v", err)
	}
	return &ret, nil
}

func (i *imageBiz) UpdateTags(ctx context.Context, userUUID string, imageUUID string, r *api.UpdateImageTagsRequest) error {
	defer log.FuncEntryWithContext(ctx, userUUID, imageUUID, r)()
	// 参数验证
	if !govalidator.IsUUID(imageUUID) {
		err := errno.ErrInvalidParameter.SetMessage("invalid image UUID")
		log.ErrorWithFunc(err, "参数验证失败", "imageUUID", imageUUID)
		return err
	}
	if !govalidator.IsUUID(userUUID) {
		return errno.ErrInvalidParameter.SetMessage("invalid user UUID")
	}
	if len(r.Tags) == 0 {
		return errno.ErrInvalidParameter.SetMessage("tags cannot be empty")
	}

	// 获取图片信息
	imageM, err := i.db.Image().Get(ctx, imageUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := errno.ErrImageNotFound.SetMessage("image %s not found", imageUUID)
			log.ErrorWithFunc(err, "图片不存在", "imageUUID", imageUUID)
			return err
		}
		log.ErrorWithFunc(err, "获取图片信息失败", "imageUUID", imageUUID)
		return errno.InternalServerError.SetMessage("failed to get image: %v", err)
	}

	// 权限验证
	if imageM.UserUUID.String() != userUUID {
		err := errno.ErrUnauthorized.SetMessage("access denied for image %s", imageUUID)
		log.ErrorWithFunc(err, "用户无权限操作该图片",
			"requestUserUUID", userUUID,
			"imageOwnerUUID", imageM.UserUUID.String())
		return err
	}

	// 标签去重处理
	tagSet := make(map[string]struct{})
	var uniqueTags []string
	for _, tag := range r.Tags {
		normalized := strings.TrimSpace(tag)
		if _, ok := tagSet[normalized]; !ok {
			tagSet[tag] = struct{}{}
			uniqueTags = append(uniqueTags, normalized)
		}
	}
	if len(uniqueTags) > config.GetImage().MaxTagCount {
		err := errno.ErrInvalidParameter.SetMessage("Image's tags too much, max count is %d,action %d", config.GetImage().MaxTagCount, len(uniqueTags))
		log.ErrorWithFunc(err, "图片标签过多", "parameter", "request")
		return err
	}

	// 更新标签
	if err := i.db.Image().AddTagsToImage(ctx, imageUUID, uniqueTags); err != nil {
		return errno.InternalServerError.SetMessage("failed to update image tags: %v", err)
	}
	return nil
}

func (i *imageBiz) Delete(ctx context.Context, userUUID string, imageUUID string) error {
	defer log.FuncEntryWithContext(ctx, userUUID, imageUUID)()
	// 参数验证
	if !govalidator.IsUUID(imageUUID) {
		return errno.ErrInvalidParameter.SetMessage("invalid image UUID")
	}
	if !govalidator.IsUUID(userUUID) {
		return errno.ErrInvalidParameter.SetMessage("invalid user UUID")
	}

	// 获取图片信息
	imageM, err := i.db.Image().Get(ctx, imageUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrImageNotFound.SetMessage("image %s not found", imageUUID)
		}
		return errno.InternalServerError.SetMessage("failed to get image: %v", err)
	}

	// 权限验证
	if imageM.UserUUID.String() != userUUID {
		return errno.ErrUnauthorized.SetMessage("access denied for image %s", imageUUID)
	}

	// 删除图片
	if err := i.db.Image().Delete(ctx, imageUUID); err != nil {
		return errno.InternalServerError.SetMessage("failed to delete image: %v", err)
	}
	return nil
}

func (i *imageBiz) DeleteCollection(ctx context.Context, userUUID string, imageUUIDs []string) error {
	//TODO implement me
	panic("implement me")
}

func (i *imageBiz) Get(ctx context.Context, userUUID string, imageUUID string) (*api.GetImageInfoResponse, error) {
	defer log.FuncEntryWithContext(ctx, userUUID, imageUUID)()
	isAnonymous := userUUID == ""

	// 参数验证
	if !govalidator.IsUUID(imageUUID) {
		return nil, errno.ErrInvalidParameter.SetMessage("invalid image UUID")
	}
	if !isAnonymous && !govalidator.IsUUID(userUUID) {
		return nil, errno.ErrInvalidParameter.SetMessage("invalid user UUID")
	}

	// 获取图片信息
	imageM, err := i.db.Image().Get(ctx, imageUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrImageNotFound.SetMessage("image %s not found", imageUUID)
		}
		return nil, errno.InternalServerError.SetMessage("failed to get image: %v", err)
	}

	// 权限验证
	if !imageM.IsPublic {
		if isAnonymous {
			return nil, errno.ErrUnauthorized.SetMessage("authentication required")
		}
		if imageM.UserUUID.String() != userUUID {
			return nil, errno.ErrUnauthorized.SetMessage("access denied for image %s", imageUUID)
		}
	}

	// 复制响应数据
	var ret api.GetImageInfoResponse
	if err := copyImageInfo((*api.ImageInfo)(&ret), imageM); err != nil {
		return nil, errno.InternalServerError.SetMessage("failed to copy image data: %v", err)
	}
	return &ret, nil
}

func (i *imageBiz) ListUserOwnImages(ctx context.Context, userUUID string, offset, limit int) (*api.ListImageResponse, error) {
	// 参数验证
	if offset < 0 {
		return nil, errno.ErrInvalidParameter.SetMessage("offset cannot be negative")
	}
	if limit < 0 {
		return nil, errno.ErrInvalidParameter.SetMessage("limit cannot be negative")
	}

	// 获取用户图片列表
	count, imageList, err := i.db.Image().GetUserImages(ctx, userUUID, offset, limit)
	if err != nil {
		return nil, errno.InternalServerError.SetMessage("failed to get user images: %v", err)
	}

	// 权限验证
	if count > 0 && imageList[0].UserUUID.String() != userUUID {
		return nil, errno.ErrUnauthorized.SetMessage("unauthorized operation")
	}

	// 处理空结果
	if count == 0 {
		return &api.ListImageResponse{}, nil
	}

	// 构建响应
	var ret api.ListImageResponse
	ret.Count = int(count)
	imageInfos := make([]api.ImageInfo, len(imageList))
	for i, image := range imageList {
		if err := copyImageInfo(&imageInfos[i], image); err != nil {
			return nil, errno.InternalServerError.SetMessage("failed to copy image data: %v", err)
		}
	}
	ret.ImageList = imageInfos
	return &ret, nil
}

func (i *imageBiz) ListUserOwnPublicImages(ctx context.Context, userUUID string, offset, limit int) (*api.ListImageResponse, error) {
	// 参数验证
	if offset < 0 {
		return nil, errno.ErrInvalidParameter.SetMessage("offset cannot be negative")
	}
	if limit < 0 {
		return nil, errno.ErrInvalidParameter.SetMessage("limit cannot be negative")
	}

	// 获取用户图片列表
	count, imageList, err := i.db.Image().GetUserImages(ctx, userUUID, offset, limit)
	if err != nil {
		return nil, errno.InternalServerError.SetMessage("failed to get user images: %v", err)
	}

	// 构建响应，只包含公开图片
	var ret api.ListImageResponse
	ret.Count = int(count)
	imageInfos := make([]api.ImageInfo, 0)
	for _, image := range imageList {
		if !image.IsPublic {
			continue
		}
		var tempImage api.ImageInfo
		if err := copyImageInfo(&tempImage, image); err != nil {
			return nil, errno.InternalServerError.SetMessage("failed to copy image data: %v", err)
		}
		imageInfos = append(imageInfos, tempImage)
	}
	ret.ImageList = imageInfos
	return &ret, nil
}

func (i *imageBiz) ListRandomPublicImages(ctx context.Context, limit int) (*api.ListImageResponse, error) {
	// 参数验证
	if limit < 0 || limit > config.GetGetImage().RandomPublicLimit {
		return nil, errno.ErrInvalidParameter.SetMessage("limit cannot be negative")
	}

	// 获取随机公开图片列表
	count, imageList, err := i.db.Image().GetRandomPublicImages(ctx, limit)
	if err != nil {
		return nil, errno.InternalServerError.SetMessage("failed to get random public images: %v", err)
	}

	// 构建响应
	var ret api.ListImageResponse
	ret.Count = int(count)
	imageInfos := make([]api.ImageInfo, len(imageList))
	for i, image := range imageList {
		if err := copyImageInfo(&imageInfos[i], image); err != nil {
			return nil, errno.InternalServerError.SetMessage("failed to copy image data: %v", err)
		}
	}
	ret.ImageList = imageInfos
	return &ret, nil
}

func (i *imageBiz) GetImageFile(ctx context.Context, userUUID string, imageUUIDFileName string) (filePath string, err error) {
	isAnonymous := userUUID == ""

	// 解析文件名和扩展名
	ext := filepath.Ext(imageUUIDFileName)
	imageUUID := strings.TrimSuffix(imageUUIDFileName, ext)
	if ext == "" || len(ext) >= len(imageUUIDFileName) {
		return "", errno.ErrImageFileInvalid.SetMessage("missing or invalid file extension")
	}
	if !govalidator.IsUUID(imageUUID) {
		return "", errno.ErrInvalidParameter.SetMessage("invalid image UUID")
	}

	// 验证文件格式
	switch ext {
	case ".png", ".webp", ".avif", ".jpg", ".jpeg":
		// 支持的格式
	default:
		return "", errno.ErrImageFileInvalid.SetMessage("unsupported image format %s", ext)
	}

	// 参数验证
	if !isAnonymous && !govalidator.IsUUID(userUUID) {
		return "", errno.ErrInvalidParameter.SetMessage("invalid user UUID")
	}

	// 获取图片信息
	imageM, err := i.db.Image().Get(ctx, imageUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errno.ErrImageNotFound.SetMessage("image %s not found", imageUUID)
		}
		return "", errno.InternalServerError.SetMessage("failed to get image: %v", err)
	}

	// 权限验证
	if !imageM.IsPublic {
		if userUUID == "" {
			return "", errno.ErrUnauthorized.SetMessage("authentication required")
		}
		if imageM.UserUUID.String() != userUUID {
			return "", errno.ErrUnauthorized.SetMessage("access denied for image %s", imageUUID)
		}
	}

	// 获取文件路径
	filePath, err = i.imageFileStore.GetFilePath(string(imageM.Hash), ext)
	if err != nil {
		return "", errno.InternalServerError.SetMessage("failed to find file: %v", err)
	}
	return filePath, nil
}
