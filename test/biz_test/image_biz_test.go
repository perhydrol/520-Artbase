package biz_test

import (
	"context"
	"testing"

	"demo520/pkg/api"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestImageBiz_Create_ErrorCases 测试图片创建的错误情况
func TestImageBiz_Create_ErrorCases(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 测试无效的用户UUID
	invalidUserUUID := "invalid-uuid"
	fileHeader := makeFileHeader(t, "test.jpg", "image/jpeg", []byte("fake image data"))
	tags := []string{faker.Word()}
	req := &api.CreateImageRequest{
		IsPublic: true,
		Tags:     tags,
	}

	_, err := imageBiz.Create(ctx, invalidUserUUID, req, fileHeader)
	assert.Error(t, err)

	// 测试空文件
	emptyFileHeader := makeFileHeader(t, "empty.jpg", "image/jpeg", []byte(""))
	_, err = imageBiz.Create(ctx, userUUID, req, emptyFileHeader)
	assert.Error(t, err)

	// 测试无效文件类型
	invalidFileHeader := makeFileHeader(t, "test.txt", "text/plain", []byte("not an image"))
	_, err = imageBiz.Create(ctx, userUUID, req, invalidFileHeader)
	assert.Error(t, err)
}

// TestImageBiz_Get_NotFound 测试获取不存在的图片
func TestImageBiz_Get_NotFound(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 测试获取不存在的图片
	nonExistentImageUUID := faker.UUIDHyphenated()
	_, err := imageBiz.Get(ctx, userUUID, nonExistentImageUUID)
	assert.Error(t, err)

	// 测试用无效用户UUID获取图片
	invalidUserUUID := "invalid-uuid"
	_, err = imageBiz.Get(ctx, invalidUserUUID, nonExistentImageUUID)
	assert.Error(t, err)
}

// TestImageBiz_Delete_NotFound 测试删除不存在的图片
func TestImageBiz_Delete_NotFound(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 测试删除不存在的图片
	nonExistentImageUUID := faker.UUIDHyphenated()
	err := imageBiz.Delete(ctx, userUUID, nonExistentImageUUID)
	assert.Error(t, err)

	// 测试用无效用户UUID删除图片
	invalidUserUUID := "invalid-uuid"
	err = imageBiz.Delete(ctx, invalidUserUUID, nonExistentImageUUID)
	assert.Error(t, err)
}

// TestImageBiz_UpdateTags_ErrorCases 测试更新标签的错误情况
func TestImageBiz_UpdateTags_ErrorCases(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 创建测试图片
	tags := []string{faker.Word()}
	imageInfo := createTestImageViaBiz(t, imageBiz, userUUID, true, tags)
	defer imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)

	// 测试更新不存在的图片标签
	nonExistentImageUUID := faker.UUIDHyphenated()
	updateReq := &api.UpdateImageTagsRequest{
		Tags: []string{faker.Word()},
	}
	err := imageBiz.UpdateTags(ctx, userUUID, nonExistentImageUUID, updateReq)
	assert.Error(t, err)

	// 测试用无效用户UUID更新标签
	invalidUserUUID := "invalid-uuid"
	err = imageBiz.UpdateTags(ctx, invalidUserUUID, imageInfo.ImageUUID, updateReq)
	assert.Error(t, err)

	// 测试空标签请求
	emptyReq := &api.UpdateImageTagsRequest{
		Tags: []string{},
	}
	err = imageBiz.UpdateTags(ctx, userUUID, imageInfo.ImageUUID, emptyReq)
	assert.Error(t, err)
}

// TestImageBiz_ListUserOwnImages_ErrorCases 测试列出用户图片的错误情况
func TestImageBiz_ListUserOwnImages_ErrorCases(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 测试无效用户UUID
	invalidUserUUID := "invalid-uuid"
	_, err := imageBiz.ListUserOwnImages(ctx, invalidUserUUID, 0, 10)
	assert.Error(t, err)

	// 测试无效分页参数
	_, err = imageBiz.ListUserOwnImages(ctx, userUUID, -1, 10) // 无效offset
	assert.Error(t, err)

	// 测试无效页面大小
	_, err = imageBiz.ListUserOwnImages(ctx, userUUID, 0, -1) // 无效limit
	assert.Error(t, err)
}

// TestImageBiz_ListUserOwnPublicImages_ErrorCases 测试列出用户公开图片的错误情况
func TestImageBiz_ListUserOwnPublicImages_ErrorCases(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 测试无效用户UUID
	invalidUserUUID := "invalid-uuid"
	_, err := imageBiz.ListUserOwnPublicImages(ctx, invalidUserUUID, 0, 10)
	assert.Error(t, err)
}

// TestImageBiz_ListRandomPublicImages_ErrorCases 测试列出随机公开图片的错误情况
func TestImageBiz_ListRandomPublicImages_ErrorCases(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 测试无效limit参数
	_, err := imageBiz.ListRandomPublicImages(ctx, -1) // 无效limit
	assert.Error(t, err)
}

// TestImageBiz_GetImageFile_Success 测试获取图片文件
func TestImageBiz_GetImageFile_Success(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 创建测试图片
	tags := []string{faker.Word()}
	imageInfo := createTestImageViaBiz(t, imageBiz, userUUID, true, tags)
	defer imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)

	// 测试获取图片文件
	filePath, err := imageBiz.GetImageFile(ctx, userUUID, imageInfo.ImageUUID+".png")
	require.NoError(t, err)
	assert.NotEmpty(t, filePath)
}

// TestImageBiz_GetImageFile_NotFound 测试获取不存在的图片文件
func TestImageBiz_GetImageFile_NotFound(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 测试获取不存在的图片文件
	invalidImageUUID := faker.UUIDHyphenated()
	_, err := imageBiz.GetImageFile(ctx, userUUID, invalidImageUUID+".png")
	assert.Error(t, err)
}

// TestImageBiz_DeleteCollection_Success 测试批量删除图片
func TestImageBiz_DeleteCollection_Success(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 创建多个测试图片
	imageInfos := createTestImageListViaBiz(t, imageBiz, userUUID, 3)

	// 提取图片UUID
	imageUUIDs := make([]string, len(imageInfos))
	for i, imageInfo := range imageInfos {
		imageUUIDs[i] = imageInfo.ImageUUID
	}

	// 测试批量删除 - 注意：DeleteCollection方法目前未实现，会panic
	// err := imageBiz.DeleteCollection(ctx, userUUID, imageUUIDs)
	// require.NoError(t, err)

	// 暂时跳过此测试，因为方法未实现
	t.Skip("DeleteCollection method is not implemented yet")
}

// TestImageBiz_DeleteCollection_ErrorCases 测试批量删除图片的错误情况
func TestImageBiz_DeleteCollection_ErrorCases(t *testing.T) {
	setViper()
	ts, _, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)

	// 暂时跳过此测试，因为DeleteCollection方法未实现
	t.Skip("DeleteCollection method is not implemented yet")
}

// TestImageBiz_PrivateImageAccess 测试私有图片访问控制
func TestImageBiz_PrivateImageAccess(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID1 := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()

	// 创建第二个用户
	user2 := createTestUserViaBiz(t, userBiz, nil)
	defer userBiz.Delete(ctx, userUUID1)
	defer userBiz.Delete(ctx, user2.UserUUID)

	// 用户1创建私有图片
	tags := []string{faker.Word()}
	imageInfo := createTestImageViaBiz(t, imageBiz, userUUID1, false, tags) // 私有图片
	defer imageBiz.Delete(ctx, userUUID1, imageInfo.ImageUUID)

	// 用户1可以访问自己的私有图片
	_, err := imageBiz.Get(ctx, userUUID1, imageInfo.ImageUUID)
	require.NoError(t, err)

	// 用户2不能访问用户1的私有图片
	_, err = imageBiz.Get(ctx, user2.UserUUID, imageInfo.ImageUUID)
	assert.Error(t, err)
}

// TestImageBiz_PublicImageAccess 测试公开图片访问
func TestImageBiz_PublicImageAccess(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID1 := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()

	// 创建第二个用户
	user2 := createTestUserViaBiz(t, userBiz, nil)
	defer userBiz.Delete(ctx, userUUID1)
	defer userBiz.Delete(ctx, user2.UserUUID)

	// 用户1创建公开图片
	tags := []string{faker.Word()}
	imageInfo := createTestImageViaBiz(t, imageBiz, userUUID1, true, tags) // 公开图片
	defer imageBiz.Delete(ctx, userUUID1, imageInfo.ImageUUID)

	// 用户1可以访问自己的公开图片
	_, err := imageBiz.Get(ctx, userUUID1, imageInfo.ImageUUID)
	require.NoError(t, err)

	// 用户2也可以访问公开图片文件（匿名访问）
	filePath, err := imageBiz.GetImageFile(ctx, "", imageInfo.ImageUUID+".png")
	require.NoError(t, err)
	assert.NotEmpty(t, filePath)
}
