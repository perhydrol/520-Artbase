package controller_test

import (
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"demo520/pkg/api"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var test_image_path = "../test_image.png"
var test_iamge_list_path = "../test_image"

func setupImageDatabase() (*gorm.DB, error) {
	// 3. 构造 DSN
	dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3316)/testdb?charset=utf8mb4&parseTime=True&loc=Local", "testpassword")

	// 4. 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.UserM{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&model.ImageM{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&model.ImageTagM{}); err != nil {
		return nil, err
	}
	return db, nil
}

func genUser(db *gorm.DB) (*api.UserInfo, string, error) {
	createUserReq := genCreateUserReq()
	userController := getUserController(db)
	c, _ := createTestContext("POST", "/users", &createUserReq)
	userController.Create(c)

	c, w := genGetUserReq(createUserReq.Email)
	userController.Get(c)
	var getResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &getResp); err != nil {
		return nil, "", err
	}
	userInfo := api.UserInfo{
		Email:    getResp["email"],
		UserUUID: getResp["user_uuid"],
		Nickname: getResp["nickname"],
		CreateAt: getResp["create_at"],
	}
	return &userInfo, createUserReq.Password, nil
}

func genImages(t *testing.T, userUUID, userToken string, db *gorm.DB) []*api.ImageInfo {
	entries, err := os.ReadDir(test_iamge_list_path)
	require.NoError(t, err)
	imageInfoList := make([]*api.ImageInfo, len(entries))
	wg := sync.WaitGroup{}
	for i, e := range entries {
		wg.Add(1)
		go func() {
			defer wg.Done()
			createImageReq := api.CreateImageRequest{
				UserUUID: userUUID,
				IsPublic: rand.Intn(100)%2 == 0,
				Tags:     []string{faker.Word(), faker.Word()},
			}
			c, w := prepareContextWithFile(t, filepath.Join(test_iamge_list_path, e.Name()), &createImageReq)
			appendJWTHeader(c, userToken)

			imageController := getImageController(db)
			imageController.Create(c)
			assert.Equal(t, http.StatusOK, w.Code)
			var imageInfo api.ImageInfo
			if err := json.Unmarshal(w.Body.Bytes(), &imageInfo); err != nil {
				t.Errorf("failed to get image create Resp: %v", err)
				return
			}
			imageInfoList[i] = &imageInfo
		}()
	}
	wg.Wait()
	return imageInfoList
}

func TestImage_Create_Success(t *testing.T) {
	setViper()
	defer cleanTestData()
	log.Init(nil)
	db, err := setupImageDatabase()
	require.NoError(t, err)
	user, password, err := genUser(db)
	require.NoError(t, err)

	userToken, err := loginAndGetToken(db, user.Email, password)
	require.NoError(t, err)

	createImageReq := api.CreateImageRequest{
		UserUUID: user.UserUUID,
		IsPublic: true,
		Tags:     []string{faker.Word(), faker.Word()},
	}
	c, w := prepareContextWithFile(t, test_image_path, &createImageReq)
	appendJWTHeader(c, userToken)

	imageController := getImageController(db)
	imageController.Create(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestImage_CreatePublicAndNoLoginGet_Success(t *testing.T) {
	setViper()
	defer cleanTestData()
	log.Init(nil)
	db, err := setupImageDatabase()
	require.NoError(t, err)
	user, password, err := genUser(db)
	require.NoError(t, err)

	userToken, err := loginAndGetToken(db, user.Email, password)
	require.NoError(t, err)

	createImageReq := api.CreateImageRequest{
		UserUUID: user.UserUUID,
		IsPublic: true,
		Tags:     []string{faker.Word(), faker.Word()},
	}
	c, w := prepareContextWithFile(t, test_image_path, &createImageReq)
	appendJWTHeader(c, userToken)

	imageController := getImageController(db)
	imageController.Create(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var imageInfo api.ImageInfo
	if err := json.Unmarshal(w.Body.Bytes(), &imageInfo); err != nil {
		t.Fatalf("failed to get image create Resp: %v", err)
	}

	c, w = createTestContext("GET", "/image", nil)
	c.Params = []gin.Param{
		{Key: "imageuuid", Value: imageInfo.ImageUUID},
	}
	imageController.Get(c)
	var getRet api.ImageInfo
	if err := json.Unmarshal(w.Body.Bytes(), &getRet); err != nil {
		t.Fatalf("failed to get image create Resp: %v", err)
	}

	assert.Equal(t, getRet.UserUUID, imageInfo.UserUUID)
	assert.Equal(t, getRet.ImageUUID, imageInfo.ImageUUID)
	assert.Equal(t, getRet.IsPublic, imageInfo.IsPublic)
}

func TestImage_CreateAndNoLoginGet_Failed(t *testing.T) {
	setViper()
	defer cleanTestData()
	log.Init(nil)
	db, err := setupImageDatabase()
	require.NoError(t, err)
	user, password, err := genUser(db)
	require.NoError(t, err)

	userToken, err := loginAndGetToken(db, user.Email, password)
	require.NoError(t, err)

	createImageReq := api.CreateImageRequest{
		UserUUID: user.UserUUID,
		IsPublic: false,
		Tags:     []string{faker.Word(), faker.Word()},
	}
	c, w := prepareContextWithFile(t, test_image_path, &createImageReq)
	appendJWTHeader(c, userToken)

	imageController := getImageController(db)
	imageController.Create(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var imageInfo api.ImageInfo
	if err := json.Unmarshal(w.Body.Bytes(), &imageInfo); err != nil {
		t.Fatalf("failed to get image create Resp: %v", err)
	}

	c, w = createTestContext("GET", "/image", nil)
	c.Params = []gin.Param{
		{Key: "imageuuid", Value: imageInfo.ImageUUID},
	}
	imageController.Get(c)
	assert.True(t, http.StatusOK != w.Code)
	var getRet api.ImageInfo
	if err := json.Unmarshal(w.Body.Bytes(), &getRet); err != nil {

	}
	assert.Equal(t, getRet.ImageUUID, "")
}

func TestImage_CreateAndLoginGet_Success(t *testing.T) {
	setViper()
	defer cleanTestData()
	log.Init(nil)
	db, err := setupImageDatabase()
	require.NoError(t, err)
	user, password, err := genUser(db)
	require.NoError(t, err)

	userToken, err := loginAndGetToken(db, user.Email, password)
	require.NoError(t, err)

	createImageReq := api.CreateImageRequest{
		UserUUID: user.UserUUID,
		IsPublic: false,
		Tags:     []string{faker.Word(), faker.Word()},
	}
	c, w := prepareContextWithFile(t, test_image_path, &createImageReq)
	appendJWTHeader(c, userToken)

	imageController := getImageController(db)
	imageController.Create(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var imageInfo api.ImageInfo
	if err := json.Unmarshal(w.Body.Bytes(), &imageInfo); err != nil {
		t.Fatalf("failed to get image create Resp: %v", err)
	}

	c, w = createTestContext("GET", "/image", nil)
	appendJWTHeader(c, userToken)
	c.Params = []gin.Param{
		{Key: "imageuuid", Value: imageInfo.ImageUUID},
	}
	imageController.Get(c)
	var getRet api.ImageInfo
	if err := json.Unmarshal(w.Body.Bytes(), &getRet); err != nil {
		t.Fatalf("failed to get image create Resp: %v", err)
	}

	assert.Equal(t, getRet.UserUUID, imageInfo.UserUUID)
	assert.Equal(t, getRet.ImageUUID, imageInfo.ImageUUID)
	assert.Equal(t, getRet.IsPublic, imageInfo.IsPublic)
}

func TestImage_CreateImagesAndGetList_Success(t *testing.T) {
	setViper()
	defer cleanTestData()
	log.Init(nil)
	db, err := setupImageDatabase()
	require.NoError(t, err)
	user, password, err := genUser(db)
	require.NoError(t, err)

	userToken, err := loginAndGetToken(db, user.Email, password)
	require.NoError(t, err)

	imageController := getImageController(db)
	imageInfoList := genImages(t, user.UserUUID, userToken, db)

	for _, imageInfo := range imageInfoList {
		c, w := createTestContext("GET", "/image", nil)
		appendJWTHeader(c, userToken)
		c.Params = []gin.Param{
			{Key: "imageuuid", Value: imageInfo.ImageUUID},
		}
		imageController.Get(c)
		var getRet api.ImageInfo
		if err := json.Unmarshal(w.Body.Bytes(), &getRet); err != nil {
			t.Fatalf("failed to get image create Resp: %v", err)
		}

		assert.Equal(t, getRet.UserUUID, imageInfo.UserUUID)
		assert.Equal(t, getRet.ImageUUID, imageInfo.ImageUUID)
		assert.Equal(t, getRet.IsPublic, imageInfo.IsPublic)
	}
}
