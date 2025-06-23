package biz_test

import (
	"bytes"
	"context"
	"demo520/internal/520/biz"
	"demo520/internal/520/biz/image"
	"demo520/internal/520/biz/user"
	"demo520/internal/520/store"
	"demo520/internal/pkg/config"
	"demo520/pkg/api"
	"demo520/test/testhelper"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"math/rand"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var test_image_path = "../test_image/cg_tm01_0101__base_cg_tm01_0101(CUnet)(noise_scale)(Level1)(x2).png"
var test_iamge_list_path = "../test_image"

// setupImageBizTest 设置图片biz层测试环境
func setupImageBizTest(t *testing.T) (*testhelper.TestSuite, image.ImageBiz, user.UserBiz, string) {
	ts := testhelper.NewTestSuite(t, nil)
	iStore := store.NewStore(ts.DB)
	bizInstance := biz.NewIBiz(iStore)
	imageBiz := bizInstance.Images()
	userBiz := bizInstance.Users()

	// 创建测试用户
	userReq := &api.CreateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
		Password: faker.Password(),
	}
	ctx := context.Background()
	err := userBiz.Create(ctx, userReq)
	require.NoError(t, err)

	userInfo, err := userBiz.Get(ctx, userReq.Email)
	require.NoError(t, err)

	return ts, imageBiz, userBiz, userInfo.UserUUID
}

// teardownImageBizTest 清理图片biz层测试环境
func teardownImageBizTest(t *testing.T, ts *testhelper.TestSuite) {
	ts.Cleanup(t)
}

// createTestImageViaBiz 通过biz层创建测试图片
func createTestImageViaBiz(t *testing.T, imageBiz image.ImageBiz, userUUID string, isPublic bool, tags []string) *api.ImageInfo {
	imageByte, err := os.ReadFile(test_image_path)
	require.NoError(t, err)
	fileHeader := makeFileHeader(t, filepath.Base(test_image_path), "image/png", imageByte)
	ctx := context.Background()
	req := &api.CreateImageRequest{
		UserUUID: userUUID,
		IsPublic: isPublic,
		Tags:     tags,
	}
	resp, err := imageBiz.Create(ctx, userUUID, req, fileHeader)
	require.NoError(t, err)
	return (*api.ImageInfo)(resp)
}

func makeFileHeader(t *testing.T, filename, contentType string, content []byte) *multipart.FileHeader {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/", buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	mf, fh, err := req.FormFile("file")
	require.NoError(t, err)
	defer mf.Close()

	return fh
}

func setViper() {
	// 使用配置管理包初始化配置
	config.Init("")
}

func cleanTestData() {
	entries, err := os.ReadDir("temp_image")
	if err != nil {
		return
	}
	for _, e := range entries {
		err := os.RemoveAll(filepath.Join("temp_image", e.Name()))
		if err != nil {
			return
		}
	}
}

// createTestImageListViaBiz 通过biz层创建测试图片列表
func createTestImageListViaBiz(t *testing.T, imageBiz image.ImageBiz, userUUID string, count int) []*api.CreateImageResponse {
	ctx := context.Background()
	if _, err := os.Stat(test_image_path); os.IsNotExist(err) {
		p, _ := os.Getwd()
		t.Fatalf("测试资源路径错误，当前工作目录：%v", p)
	}

	imageList := make([]*api.CreateImageResponse, count)
	for i := 0; i < count; i++ {
		imageByte, err := os.ReadFile(test_image_path)
		require.NoError(t, err)
		fileHeader := makeFileHeader(t, fmt.Sprintf("test_image_%d.png", i), "image/png", imageByte)
		createImageReq := api.CreateImageRequest{
			UserUUID: userUUID,
			IsPublic: rand.Intn(10)%2 == 0, // 随机设置公开状态
			Tags:     []string{faker.Word(), faker.Word()},
		}
		imageList[i], err = imageBiz.Create(ctx, userUUID, &createImageReq, fileHeader)
		require.NoError(t, err)
	}
	return imageList
}

func TestImage_Create_Success(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	tags := []string{faker.Word(), faker.Word()}
	imageByte, err := os.ReadFile(test_image_path)
	require.NoError(t, err)
	fileHeader := makeFileHeader(t, filepath.Base(test_image_path), "image/png", imageByte)
	req := &api.CreateImageRequest{
		UserUUID: userUUID,
		IsPublic: true,
		Tags:     tags,
	}

	ctx := context.Background()
	resp, err := imageBiz.Create(ctx, userUUID, req, fileHeader)
	require.NoError(t, err)

	imageInfo := (*api.ImageInfo)(resp)
	defer imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)

	assert.Equal(t, userUUID, imageInfo.UserUUID)
	assert.True(t, imageInfo.IsPublic)
	assert.Equal(t, tags, imageInfo.Tags)
	assert.NotEmpty(t, imageInfo.ImageUUID)
	// assert.NotEmpty(t, imageInfo.Token)

	// 清理用户
	userBiz.Delete(ctx, userUUID)
}

func TestImage_Del_Success(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	tags := []string{faker.Word(), faker.Word()}
	imageInfo := createTestImageViaBiz(t, imageBiz, userUUID, true, tags)

	ctx := context.Background()
	err := imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)
	require.NoError(t, err)

	// 验证图片已被删除
	_, err = imageBiz.Get(ctx, userUUID, imageInfo.ImageUUID)
	assert.Error(t, err)

	// 清理用户
	userBiz.Delete(ctx, userUUID)
}

func TestImage_Get_Success(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	tags := []string{faker.Word(), faker.Word()}
	imageInfo := createTestImageViaBiz(t, imageBiz, userUUID, true, tags)

	ctx := context.Background()
	defer imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)

	getInfo, err := imageBiz.Get(ctx, userUUID, imageInfo.ImageUUID)
	require.NoError(t, err)

	assert.Equal(t, imageInfo.Tags, getInfo.Tags)
	// assert.Equal(t, imageInfo.Token, getInfo.Token)
	assert.Equal(t, imageInfo.UserUUID, getInfo.UserUUID)
	assert.Equal(t, imageInfo.IsPublic, getInfo.IsPublic)
	assert.Equal(t, imageInfo.ImageUUID, getInfo.ImageUUID)
	assert.Equal(t, imageInfo.CreatedAt, getInfo.CreatedAt)
	assert.Equal(t, imageInfo.UpdatedAt, getInfo.UpdatedAt)

	// 清理用户
	userBiz.Delete(ctx, userUUID)
}

func TestImage_UpdateTags(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	initialTags := []string{faker.Word(), faker.Word()}
	imageInfo := createTestImageViaBiz(t, imageBiz, userUUID, true, initialTags)

	ctx := context.Background()
	defer imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)

	updateTags := []string{faker.Word(), faker.Word()}
	updateReq := api.UpdateImageTagsRequest{
		Tags: updateTags,
	}

	err := imageBiz.UpdateTags(ctx, imageInfo.UserUUID, imageInfo.ImageUUID, &updateReq)
	require.NoError(t, err)

	getInfo, err := imageBiz.Get(ctx, userUUID, imageInfo.ImageUUID)
	require.NoError(t, err)

	// 验证标签已更新（根据实际业务逻辑，可能是追加或替换）
	expectedTags := make([]string, len(imageInfo.Tags)+len(updateTags))
	copy(expectedTags, imageInfo.Tags)
	copy(expectedTags[len(imageInfo.Tags):], updateTags)
	assert.Equal(t, expectedTags, getInfo.Tags)

	// 验证其他字段未变
	// assert.Equal(t, imageInfo.Token, getInfo.Token)
	assert.Equal(t, imageInfo.UserUUID, getInfo.UserUUID)
	assert.Equal(t, imageInfo.IsPublic, getInfo.IsPublic)
	assert.Equal(t, imageInfo.ImageUUID, getInfo.ImageUUID)
	assert.Equal(t, imageInfo.CreatedAt, getInfo.CreatedAt)

	// 清理用户
	userBiz.Delete(ctx, userUUID)
}

func TestImage_ListUserOwnImages(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	imageInfos := createTestImageListViaBiz(t, imageBiz, userUUID, 5)

	defer func() {
		for _, imageInfo := range imageInfos {
			imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)
		}
		userBiz.Delete(ctx, userUUID)
	}()

	listResp, err := imageBiz.ListUserOwnImages(ctx, userUUID, 0, 10)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(listResp.ImageList), len(imageInfos))

	// 验证返回的图片都属于该用户
	for _, imageInfo := range listResp.ImageList {
		assert.Equal(t, userUUID, imageInfo.UserUUID)
	}
}

func TestImage_ListUserOwnPublicImages(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	imageInfos := createTestImageListViaBiz(t, imageBiz, userUUID, 5)

	defer func() {
		for _, imageInfo := range imageInfos {
			imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)
		}
		userBiz.Delete(ctx, userUUID)
	}()

	listResp, err := imageBiz.ListUserOwnPublicImages(ctx, userUUID, 0, 10)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(listResp.ImageList), len(imageInfos))

	// 验证返回的图片都是公开的且属于该用户
	for _, imageInfo := range listResp.ImageList {
		assert.Equal(t, userUUID, imageInfo.UserUUID)
		assert.True(t, imageInfo.IsPublic)
	}
}

func TestImage_ListRandomPublicImages(t *testing.T) {
	setViper()
	ts, imageBiz, userBiz, userUUID := setupImageBizTest(t)
	defer teardownImageBizTest(t, ts)
	defer cleanTestData()

	ctx := context.Background()
	imageInfos := createTestImageListViaBiz(t, imageBiz, userUUID, 5)

	defer func() {
		for _, imageInfo := range imageInfos {
			imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)
		}
		userBiz.Delete(ctx, userUUID)
	}()

	listResp, err := imageBiz.ListRandomPublicImages(ctx, 10)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(listResp.ImageList), len(imageInfos))

	// 验证返回的图片都是公开的
	for _, imageInfo := range listResp.ImageList {
		assert.True(t, imageInfo.IsPublic)
	}
}
