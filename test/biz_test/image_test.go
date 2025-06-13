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
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var test_image_path = "../test_image.png"

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
