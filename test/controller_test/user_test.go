package controller_test

import (
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"demo520/pkg/api"
	"demo520/pkg/token"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupUserDatabase() (*gorm.DB, error) {
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

	return db, nil
}

func genCreateUserReq() *api.CreateUserRequest {
	return &api.CreateUserRequest{
		Nickname: faker.Name(),
		Email:    faker.Email(),
		Password: faker.Password(),
	}
}

func genLoginReq(email, password string) *api.LoginRequest {
	return &api.LoginRequest{
		Email:    email,
		Password: password,
		SeedTime: time.Now().Unix(),
	}
}

func genGetUserReq(email string) (*gin.Context, *httptest.ResponseRecorder) {
	// 创建 Gin 测试上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 设置路由参数
	c.Params = gin.Params{gin.Param{Key: "email", Value: email}}
	return c, w
}

func TestUser_Create_Success(t *testing.T) {
	log.Init(nil)
	db, err := setupUserDatabase()
	require.NoError(t, err)
	createUserReq := genCreateUserReq()
	userController := getUserController(db)
	c, w := createTestContext("POST", "/users", &createUserReq)
	userController.Create(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUser_CreateAndGet_Success(t *testing.T) {
	log.Init(nil)
	db, err := setupUserDatabase()
	require.NoError(t, err)
	createUserReq := genCreateUserReq()
	userController := getUserController(db)
	c, w := createTestContext("POST", "/users", &createUserReq)
	userController.Create(c)
	assert.Equal(t, http.StatusOK, w.Code)

	c, w = genGetUserReq(createUserReq.Email)
	userController.Get(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	require.NoError(t, err)
	assert.Equal(t, createUserReq.Email, getResp["email"])
	assert.Equal(t, createUserReq.Nickname, getResp["nickname"])
}

func TestUser_CreateAndLogin_Success(t *testing.T) {
	log.Init(nil)
	db, err := setupUserDatabase()
	require.NoError(t, err)
	createUserReq := genCreateUserReq()
	userController := getUserController(db)
	c, w := createTestContext("POST", "/users", &createUserReq)
	userController.Create(c)
	assert.Equal(t, http.StatusOK, w.Code)

	c, w = genGetUserReq(createUserReq.Email)
	userController.Get(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	require.NoError(t, err)
	assert.Equal(t, createUserReq.Email, getResp["email"])
	assert.Equal(t, createUserReq.Nickname, getResp["nickname"])

	loginReq := genLoginReq(createUserReq.Email, createUserReq.Password)
	c, w = createTestContext("POST", "/login", &loginReq)
	userController.Login(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	token_claims, err := token.ParseToken(loginResp["token"])
	require.NoError(t, err)
	userUUID := token_claims.UserUUID
	assert.Equal(t, getResp["user_uuid"], userUUID)
}

func TestUser_CreateAndChangePassword_Success(t *testing.T) {
	log.Init(nil)
	db, err := setupUserDatabase()
	require.NoError(t, err)
	createUserReq := genCreateUserReq()
	userController := getUserController(db)
	c, w := createTestContext("POST", "/users", &createUserReq)
	userController.Create(c)
	assert.Equal(t, http.StatusOK, w.Code)

	c, w = genGetUserReq(createUserReq.Email)
	userController.Get(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	require.NoError(t, err)
	assert.Equal(t, createUserReq.Email, getResp["email"])
	assert.Equal(t, createUserReq.Nickname, getResp["nickname"])

	changePassword := api.ChangePasswordRequest{
		OldPassword: createUserReq.Password,
		NewPassword: faker.Password(),
	}
	c, w = createTestContext("POST", "/change", &changePassword)
	c.Params = gin.Params{gin.Param{Key: "email", Value: createUserReq.Email}}
	userController.ChangePassword(c)
	assert.Equal(t, http.StatusOK, w.Code)

	loginReq := genLoginReq(createUserReq.Email, changePassword.NewPassword)
	c, w = createTestContext("POST", "/login", &loginReq)
	userController.Login(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	token_claims, err := token.ParseToken(loginResp["token"])
	require.NoError(t, err)
	userUUID := token_claims.UserUUID
	assert.Equal(t, getResp["user_uuid"], userUUID)
}

func TestUser_CreateAndUpdate_Success(t *testing.T) {
	log.Init(nil)
	db, err := setupUserDatabase()
	require.NoError(t, err)
	createUserReq := genCreateUserReq()
	userController := getUserController(db)
	c, w := createTestContext("POST", "/users", &createUserReq)
	userController.Create(c)
	assert.Equal(t, http.StatusOK, w.Code)

	loginReq := genLoginReq(createUserReq.Email, createUserReq.Password)
	c, w = createTestContext("POST", "/login", &loginReq)
	userController.Login(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	userToken := loginResp["token"]

	updateReq := api.UpdateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
	}
	c, w = createTestContext("POST", "/change", &updateReq)
	appendJWTHeader(c, userToken)
	c.Params = gin.Params{gin.Param{Key: "email", Value: createUserReq.Email}}
	userController.Update(c)
	assert.Equal(t, http.StatusOK, w.Code)

	c, w = genGetUserReq(updateReq.Email)
	userController.Get(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	require.NoError(t, err)
	assert.Equal(t, updateReq.Email, getResp["email"])
	assert.Equal(t, updateReq.Nickname, getResp["nickname"])
}
