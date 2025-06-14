package biz_test

import (
	"bytes"
	"context"
	"demo520/internal/520/biz"
	"demo520/internal/520/biz/image"
	"demo520/internal/520/biz/user"
	"demo520/internal/520/store"
	"demo520/internal/pkg/model"
	"demo520/pkg/api"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"math/rand"

	"github.com/go-faker/faker/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var test_image_path = "../test_image.png"
var test_iamge_list_path = "../test_image"

func setupImageDatabase() (*gorm.DB, *api.CreateUserRequest, string, error) {
	// 3. 构造 DSN
	dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3316)/testdb?charset=utf8mb4&parseTime=True&loc=Local", "testpassword")

	// 4. 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, nil, "", err
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.UserM{}); err != nil {
		return nil, nil, "", err
	}
	if err := db.AutoMigrate(&model.ImageM{}); err != nil {
		return nil, nil, "", err
	}
	if err := db.AutoMigrate(&model.ImageTagM{}); err != nil {
		return nil, nil, "", err
	}
	userReq, err := genNewUser(nil, db, nil)
	if err != nil {
		return nil, nil, "", err
	}
	ctx := context.Background()
	userBiz := getUserBiz(db)
	userInfo, err := userBiz.Get(ctx, userReq.Email)
	if err != nil {
		return nil, nil, "", err
	}
	userUUID := userInfo.UserUUID
	return db, userReq, userUUID, nil
}

func getImageBiz(db *gorm.DB) image.ImageBiz {
	iStore := store.NewStore(db)
	biz := biz.NewIBiz(iStore)
	imageBiz := biz.Images()
	return imageBiz
}

func getUserBiz(db *gorm.DB) user.UserBiz {
	iStore := store.NewStore(db)
	biz := biz.NewIBiz(iStore)
	userBiz := biz.Users()
	return userBiz
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
	viper.Set("ImageMaxSize", int64(20*1024*1024)) // 10 MB
	viper.Set("image_dir", "temp_image")
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

func create_new_image(t *testing.T, db *gorm.DB, userUUID string) *api.CreateImageResponse {
	imageBiz := getImageBiz(db)
	ctx := context.Background()
	if _, err := os.Stat(test_image_path); os.IsNotExist(err) {
		p, _ := os.Getwd()
		t.Fatalf("测试资源路径错误，当前工作目录：%v", p)
	}
	imageByte, err := os.ReadFile(test_image_path)
	require.NoError(t, err)
	fileHeader := makeFileHeader(t, "test_image.png", "image/png", imageByte)
	createImageReq := api.CreateImageRequest{
		UserUUID: userUUID,
		IsPublic: true,
		Tags:     []string{faker.Word(), faker.Word()},
	}
	createImageResp, err := imageBiz.Create(ctx, userUUID, &createImageReq, fileHeader)
	require.NoError(t, err)
	return createImageResp
}

func create_new_imageList(t *testing.T, db *gorm.DB, userUUID string) []*api.CreateImageResponse {
	imageBiz := getImageBiz(db)
	ctx := context.Background()
	if _, err := os.Stat(test_image_path); os.IsNotExist(err) {
		p, _ := os.Getwd()
		t.Fatalf("测试资源路径错误，当前工作目录：%v", p)
	}
	entries, err := os.ReadDir(test_iamge_list_path)
	require.NoError(t, err)
	imageList := make([]*api.CreateImageResponse, len(entries))
	wg := sync.WaitGroup{}
	for i, e := range entries {
		wg.Add(1)
		go func() {
			defer wg.Done()
			imageByte, err := os.ReadFile(filepath.Join(test_iamge_list_path, e.Name()))
			require.NoError(t, err)
			fileHeader := makeFileHeader(t, fmt.Sprintf("test_image_%d.png", i), "image/png", imageByte)
			createImageReq := api.CreateImageRequest{
				UserUUID: userUUID,
				IsPublic: rand.Intn(10)%2 == 0,
				Tags:     []string{faker.Word(), faker.Word()},
			}
			imageList[i], err = imageBiz.Create(ctx, userUUID, &createImageReq, fileHeader)
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	return imageList
}

func TestImage_Create_Success(t *testing.T) {
	setViper()
	db, _, userUUID, err := setupImageDatabase()
	if err != nil {
		t.Fatalf("failed to setup database: %v", err)
		return
	}
	imageBiz := getImageBiz(db)
	userBiz := getUserBiz(db)
	ctx := context.Background()
	defer userBiz.Delete(ctx, userUUID)
	if _, err := os.Stat(test_image_path); os.IsNotExist(err) {
		p, _ := os.Getwd()
		t.Fatalf("测试资源路径错误，当前工作目录：%v", p)
	}
	imageByte, err := os.ReadFile(test_image_path)
	require.NoError(t, err)
	fileHeader := makeFileHeader(t, "test_image.png", "image/png", imageByte)
	createImageReq := api.CreateImageRequest{
		UserUUID: userUUID,
		IsPublic: true,
		Tags:     []string{faker.Word(), faker.Word()},
	}
	createImageResp, err := imageBiz.Create(ctx, userUUID, &createImageReq, fileHeader)
	defer cleanTestData()
	require.NoError(t, err)
	assert.Equal(t, createImageResp.IsPublic, createImageReq.IsPublic)
	assert.Equal(t, createImageResp.UserUUID, createImageReq.UserUUID)
	assert.Equal(t, createImageResp.Tags, createImageReq.Tags)
}

func TestImage_Del_Success(t *testing.T) {
	setViper()
	db, _, userUUID, err := setupImageDatabase()
	require.NoError(t, err)
	imageInfo := create_new_image(t, db, userUUID)
	defer cleanTestData()
	imageBiz := getImageBiz(db)
	ctx := context.Background()
	err = imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)
	require.NoError(t, err)
}

func TestImage_Get_Success(t *testing.T) {
	setViper()
	db, _, userUUID, err := setupImageDatabase()
	require.NoError(t, err)
	imageBiz := getImageBiz(db)
	ctx := context.Background()
	imageInfo := create_new_image(t, db, userUUID)
	defer cleanTestData()
	defer imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)
	getInfo, err := imageBiz.Get(ctx, userUUID, imageInfo.ImageUUID)
	require.NoError(t, err)
	assert.Equal(t, imageInfo.Tags, getInfo.Tags)
	assert.Equal(t, imageInfo.Token, getInfo.Token)
	assert.Equal(t, imageInfo.UserUUID, getInfo.UserUUID)
	assert.Equal(t, imageInfo.IsPublic, getInfo.IsPublic)
	assert.Equal(t, imageInfo.ImageUUID, getInfo.ImageUUID)
	assert.Equal(t, imageInfo.CreatedAt, getInfo.CreatedAt)
	assert.Equal(t, imageInfo.UpdatedAt, getInfo.UpdatedAt)
}

func TestImage_UpdateTags(t *testing.T) {
	setViper()
	db, _, userUUID, err := setupImageDatabase()
	require.NoError(t, err)
	imageBiz := getImageBiz(db)
	ctx := context.Background()
	imageInfo := create_new_image(t, db, userUUID)
	defer cleanTestData()
	defer imageBiz.Delete(ctx, userUUID, imageInfo.ImageUUID)
	updataReq := api.UpdateImageTagsRequest{
		Tags: []string{faker.Word(), faker.Word()},
	}
	if err := imageBiz.UpdateTags(ctx, imageInfo.UserUUID, imageInfo.ImageUUID, &updataReq); err != nil {
		t.Fatalf("failed to updata tags: %v", err)
	}
	getInfo, err := imageBiz.Get(ctx, userUUID, imageInfo.ImageUUID)
	require.NoError(t, err)
	newTags := make([]string, len(imageInfo.Tags)+len(updataReq.Tags))
	copy(newTags, imageInfo.Tags)
	copy(newTags[len(imageInfo.Tags):], updataReq.Tags)
	assert.Equal(t, getInfo.Tags, newTags)
	assert.Equal(t, imageInfo.Token, getInfo.Token)
	assert.Equal(t, imageInfo.UserUUID, getInfo.UserUUID)
	assert.Equal(t, imageInfo.IsPublic, getInfo.IsPublic)
	assert.Equal(t, imageInfo.ImageUUID, getInfo.ImageUUID)
	assert.Equal(t, imageInfo.CreatedAt, getInfo.CreatedAt)
}

func TestImage_ListUserOwnImages(t *testing.T) {
	setViper()
	defer cleanTestData()
	db, _, userUUID, err := setupImageDatabase()
	require.NoError(t, err)
	imageBiz := getImageBiz(db)
	ctx := context.Background()
	imageInfo := create_new_imageList(t, db, userUUID)
	defer func() {
		for _, img := range imageInfo {
			imageBiz.Delete(ctx, userUUID, img.ImageUUID)
		}
	}()
	imageListResp, err := imageBiz.ListUserOwnImages(ctx, userUUID, 0, len(imageInfo))
	require.NoError(t, err)
	assert.Equal(t, len(imageListResp.ImageList), len(imageInfo))
	assert.Equal(t, imageListResp.Count, len(imageInfo))
	imageMap := make(map[string]*api.ImageInfo, len(imageListResp.ImageList))
	for _, img := range imageListResp.ImageList {
		imageMap[img.ImageUUID] = &img
	}
	for index := range imageListResp.ImageList {
		assert.Equal(t, imageInfo[index].Tags, imageMap[imageInfo[index].ImageUUID].Tags)
		assert.Equal(t, imageInfo[index].Token, imageMap[imageInfo[index].ImageUUID].Token)
		assert.Equal(t, imageInfo[index].UserUUID, imageMap[imageInfo[index].ImageUUID].UserUUID)
		assert.Equal(t, imageInfo[index].IsPublic, imageMap[imageInfo[index].ImageUUID].IsPublic)
		assert.Equal(t, imageInfo[index].ImageUUID, imageMap[imageInfo[index].ImageUUID].ImageUUID)
		assert.Equal(t, imageInfo[index].CreatedAt, imageMap[imageInfo[index].ImageUUID].CreatedAt)
		assert.Equal(t, imageInfo[index].UpdatedAt, imageMap[imageInfo[index].ImageUUID].UpdatedAt)
	}

	if _, err := imageBiz.ListUserOwnImages(ctx, faker.UUIDHyphenated(), 0, 10); err != nil {

	}
}

func TestImage_ListUserOwnPublicImages(t *testing.T) {
	setViper()
	defer cleanTestData()
	db, _, userUUID, err := setupImageDatabase()
	require.NoError(t, err)
	imageBiz := getImageBiz(db)
	ctx := context.Background()
	imageInfo := create_new_imageList(t, db, userUUID)
	defer func() {
		for _, img := range imageInfo {
			imageBiz.Delete(ctx, userUUID, img.ImageUUID)
		}
	}()
	public_count := 0
	for _, img := range imageInfo {
		if img.IsPublic {
			public_count++
		}
	}
	imageListResp, err := imageBiz.ListUserOwnPublicImages(ctx, userUUID, 0, len(imageInfo))
	require.NoError(t, err)
	assert.Equal(t, public_count, len(imageListResp.ImageList))
	assert.Equal(t, public_count, len(imageListResp.ImageList))
	imageMap := make(map[string]*api.ImageInfo, len(imageListResp.ImageList))
	for _, img := range imageListResp.ImageList {
		imageMap[img.ImageUUID] = &img
	}
	for index := range imageListResp.ImageList {
		assert.Equal(t, imageInfo[index].Tags, imageMap[imageInfo[index].ImageUUID].Tags)
		assert.Equal(t, imageInfo[index].Token, imageMap[imageInfo[index].ImageUUID].Token)
		assert.Equal(t, imageInfo[index].UserUUID, imageMap[imageInfo[index].ImageUUID].UserUUID)
		assert.Equal(t, imageInfo[index].IsPublic, imageMap[imageInfo[index].ImageUUID].IsPublic)
		assert.Equal(t, imageInfo[index].ImageUUID, imageMap[imageInfo[index].ImageUUID].ImageUUID)
		assert.Equal(t, imageInfo[index].CreatedAt, imageMap[imageInfo[index].ImageUUID].CreatedAt)
		assert.Equal(t, imageInfo[index].UpdatedAt, imageMap[imageInfo[index].ImageUUID].UpdatedAt)
	}

	if _, err := imageBiz.ListUserOwnPublicImages(ctx, faker.UUIDHyphenated(), 0, 10); err != nil {

	}
}

func TestImage_ListRandomPublicImages(t *testing.T) {
	setViper()
	defer cleanTestData()
	db, _, userUUID, err := setupImageDatabase()
	require.NoError(t, err)
	imageBiz := getImageBiz(db)
	ctx := context.Background()
	imageInfo := create_new_imageList(t, db, userUUID)
	defer func() {
		for _, img := range imageInfo {
			imageBiz.Delete(ctx, userUUID, img.ImageUUID)
		}
	}()
	public_count := 0
	for _, img := range imageInfo {
		if img.IsPublic {
			public_count++
		}
	}
	imageListResp, err := imageBiz.ListRandomPublicImages(ctx, len(imageInfo))
	require.NoError(t, err)
	assert.Equal(t, public_count, len(imageListResp.ImageList))
	assert.Equal(t, public_count, len(imageListResp.ImageList))
	imageMap := make(map[string]*api.ImageInfo, len(imageListResp.ImageList))
	for _, img := range imageListResp.ImageList {
		imageMap[img.ImageUUID] = &img
	}
	for _, img := range imageInfo {
		if !img.IsPublic {
			continue
		}
		img_uuid := img.ImageUUID
		assert.Equal(t, img.Tags, imageMap[img_uuid].Tags)
		assert.Equal(t, img.Token, imageMap[img_uuid].Token)
		assert.Equal(t, img.UserUUID, imageMap[img_uuid].UserUUID)
		assert.Equal(t, img.IsPublic, imageMap[img_uuid].IsPublic)
		assert.Equal(t, img.ImageUUID, imageMap[img_uuid].ImageUUID)
		assert.Equal(t, img.CreatedAt, imageMap[img_uuid].CreatedAt)
		assert.Equal(t, img.UpdatedAt, imageMap[img_uuid].UpdatedAt)
	}

	if _, err := imageBiz.ListRandomPublicImages(ctx, 10); err != nil {

	}
}
