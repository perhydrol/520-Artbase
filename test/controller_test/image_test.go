package controller_test

import (
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"demo520/pkg/api"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

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
