package controller_test

import (
	"bytes"
	"demo520/internal/520/controller/image"
	"demo520/internal/520/controller/user"
	"demo520/internal/pkg/config"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/middleware"
	"demo520/pkg/api"
	"demo520/pkg/token"
	"demo520/test/testhelper"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ControllerTestSuite controller层集成测试套件
type ControllerTestSuite struct {
	suite.Suite
	ts           *testhelper.TestSuite
	router       *gin.Engine
	userCtrl     *user.UserController
	imageCtrl    *image.ImageController
	testUser     *testhelper.TestUser
	testImageDir string
}

// SetupSuite 测试套件初始化
func (s *ControllerTestSuite) SetupSuite() {
	// 初始化配置和日志
	config.Init("../../configs/demo520.yaml")
	log.Init(log.NewLogConfig())
	token.Init("test-secret-key")

	// 设置测试数据库
	s.ts = testhelper.NewTestSuite(s.T(), nil)

	// 初始化controllers
	s.userCtrl = user.NewUserController(s.ts.Store)
	s.imageCtrl = image.NewUserController(s.ts.Store)

	// 设置gin路由
	s.setupRouter()

	// 创建测试用户
	s.createTestUser()

	// 创建测试图片目录
	s.setupTestImageDir()
}

// TearDownSuite 测试套件清理
func (s *ControllerTestSuite) TearDownSuite() {
	s.ts.Cleanup(s.T())
	// 清理测试图片目录
	if s.testImageDir != "" {
		os.RemoveAll(s.testImageDir)
	}
}

// SetupTest 每个测试前的准备
func (s *ControllerTestSuite) SetupTest() {
	// 清理数据库
	testhelper.CleanupTestDatabase(s.T(), s.ts.DB)
	// 重新创建测试用户
	s.createTestUser()
}

// setupRouter 设置测试路由
func (s *ControllerTestSuite) setupRouter() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// 注册路由
	api_v1 := s.router.Group("/v1")

	auth := api_v1.Group("/auth")
	{
		auth.POST("/register", s.userCtrl.Create)
		auth.POST("/login", s.userCtrl.Login)
		auth.POST("/change-password/:email", s.userCtrl.ChangePassword)
	}

	users := api_v1.Group("/users").Use(middleware.JWTAuth())
	{
		users.GET("/:email", s.userCtrl.Get)
		users.PATCH("/:email", s.userCtrl.Update)
	}

	images := api_v1.Group("/images")
	{
		// 公开接口
		images.GET("", s.imageCtrl.GetPublicList)
		images.GET("/user/:user_uuid/public", s.imageCtrl.GetUserPublicList)

		// 私有接口：用子路由组加 JWT
		privateImages := images.Group("").Use(middleware.JWTAuth())
		{
			privateImages.POST("", s.imageCtrl.Create)
			privateImages.GET("/:image_uuid", s.imageCtrl.Get)
			privateImages.GET("/user/:user_uuid/images", s.imageCtrl.GetUserImagesList)
			privateImages.PATCH("/:image_uuid", s.imageCtrl.UpdateImageTags)
			privateImages.GET("/file/:imageUUIDFileName", s.imageCtrl.GetImageFile)
		}
	}
}

// createTestUser 创建测试用户
func (s *ControllerTestSuite) createTestUser() {
	s.testUser = testhelper.CreateTestUser(s.T(), s.ts.DB, nil)
}

// setupTestImageDir 设置测试图片目录
func (s *ControllerTestSuite) setupTestImageDir() {
	tempDir, err := os.MkdirTemp("", "controller_test_images")
	require.NoError(s.T(), err)
	s.testImageDir = tempDir
}

// createTestImageFile 使用真实的测试图片文件
func (s *ControllerTestSuite) createTestImageFile(filename ...string) string {
	// 使用项目中的真实图片文件
	realImagePath := "/home/oxygen/project/520/test/test_image/cg_tm01_0101__base_cg_tm01_0101.png"

	// 检查文件是否存在
	if _, err := os.Stat(realImagePath); os.IsNotExist(err) {
		s.T().Fatalf("测试图片文件不存在: %s", realImagePath)
	}

	return realImagePath
}

// makeRequest 发送HTTP请求的辅助函数
func (s *ControllerTestSuite) makeRequest(method, url string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, body)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// makeJSONRequest 发送JSON请求的辅助函数
func (s *ControllerTestSuite) makeJSONRequest(method, url string, payload interface{}, token string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(payload)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	return s.makeRequest(method, url, bytes.NewReader(body), headers)
}

// generateToken 生成测试用的JWT token
func (s *ControllerTestSuite) generateToken(userUUID string) string {
	token, err := token.GenerateToken(userUUID)
	require.NoError(s.T(), err)
	return token
}

// makeMultipartRequest 发送multipart请求的辅助函数
func (s *ControllerTestSuite) makeMultipartRequest(url string, fields map[string]string, imagePath string, token string) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件
	if imagePath != "" {
		file, err := os.Open(imagePath)
		require.NoError(s.T(), err)
		defer file.Close()

		part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
		require.NoError(s.T(), err)
		_, err = io.Copy(part, file)
		require.NoError(s.T(), err)
	}

	// 添加其他字段
	for key, value := range fields {
		writer.WriteField(key, value)
	}
	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	return s.makeRequest("POST", url, body, headers)
}

// TestUserController_Create 测试用户注册
func (s *ControllerTestSuite) TestUserController_Create() {
	tests := []struct {
		name           string
		request        api.CreateUserRequest
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "成功创建用户",
			request: api.CreateUserRequest{
				Nickname: faker.Name(),
				Email:    faker.Email(),
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "邮箱格式错误",
			request: api.CreateUserRequest{
				Nickname: faker.Name(),
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "密码太短",
			request: api.CreateUserRequest{
				Nickname: faker.Name(),
				Email:    faker.Email(),
				Password: "123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := s.makeJSONRequest("POST", "/v1/auth/register", tt.request, "")
			assert.Equal(s.T(), tt.expectedStatus, w.Code)

			if !tt.expectedError {
				// 验证用户是否成功创建
				s.ts.AssertUserExists(s.T(), tt.request.Email)
			}
		})
	}
}

// TestUserController_Login 测试用户登录
func (s *ControllerTestSuite) TestUserController_Login() {
	tests := []struct {
		name           string
		request        api.LoginRequest
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "成功登录",
			request: api.LoginRequest{
				Email:    s.testUser.Email,
				Password: s.testUser.Password,
				SeedTime: time.Now().Unix(),
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "密码错误",
			request: api.LoginRequest{
				Email:    s.testUser.Email,
				Password: "wrongpassword",
				SeedTime: time.Now().Unix(),
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "用户不存在",
			request: api.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
				SeedTime: time.Now().Unix(),
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := s.makeJSONRequest("POST", "/v1/auth/login", tt.request, "")
			assert.Equal(s.T(), tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response api.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				assert.NotEmpty(s.T(), response.Token)
			}
		})
	}
}

// TestUserController_Get 测试获取用户信息
func (s *ControllerTestSuite) TestUserController_Get() {
	token := s.generateToken(s.testUser.User.UserUUID)

	tests := []struct {
		name           string
		email          string
		token          string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "成功获取用户信息",
			email:          s.testUser.Email,
			token:          token,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "未提供token",
			email:          s.testUser.Email,
			token:          "",
			expectedStatus: http.StatusOK, // 未提供Token，默认为匿名访问
			expectedError:  false,
		},
		{
			name:           "无效token",
			email:          s.testUser.Email,
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("/v1/users/%s", tt.email)
			w := s.makeJSONRequest("GET", url, nil, tt.token)
			assert.Equal(s.T(), tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response api.GetUserInfoResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), s.testUser.Email, response.Email)
				assert.Equal(s.T(), s.testUser.User.UserUUID, response.UserUUID)
			}
		})
	}
}

// TestUserController_Update 测试更新用户信息
func (s *ControllerTestSuite) TestUserController_Update() {
	token := s.generateToken(s.testUser.User.UserUUID)

	tests := []struct {
		name           string
		request        api.UpdateUserRequest
		token          string
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "成功更新用户信息",
			request: api.UpdateUserRequest{
				Email:    faker.Email(),
				Nickname: faker.Name(),
			},
			token:          token,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "邮箱格式错误",
			request: api.UpdateUserRequest{
				Email:    "invalid-email",
				Nickname: faker.Name(),
			},
			token:          token,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("/v1/users/%s", s.testUser.Email)
			w := s.makeJSONRequest("PATCH", url, tt.request, tt.token)
			assert.Equal(s.T(), tt.expectedStatus, w.Code)
		})
	}
}

// TestImageController_Create 测试图片上传
func (s *ControllerTestSuite) TestImageController_Create() {
	token := s.generateToken(s.testUser.User.UserUUID)
	testImagePath := s.createTestImageFile()

	tests := []struct {
		name           string
		token          string
		hasFile        bool
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "成功上传图片",
			token:          token,
			hasFile:        true,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "未提供token",
			token:          "",
			hasFile:        true,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "未提供文件",
			token:          token,
			hasFile:        false,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			if tt.hasFile {
				file, err := os.Open(testImagePath)
				require.NoError(s.T(), err)
				defer file.Close()

				part, err := writer.CreateFormFile("image", "test_image.jpg")
				require.NoError(s.T(), err)
				_, err = io.Copy(part, file)
				require.NoError(s.T(), err)
			}

			// 添加其他表单字段
			writer.WriteField("title", "Test Image")
			writer.WriteField("description", "Test Description")
			writer.WriteField("is_public", "true")
			writer.WriteField("tags", "test,image")
			writer.Close()

			headers := map[string]string{
				"Content-Type": writer.FormDataContentType(),
			}
			if tt.token != "" {
				headers["Authorization"] = "Bearer " + tt.token
			}

			w := s.makeRequest("POST", "/v1/images", body, headers)
			assert.Equal(s.T(), tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response api.CreateImageResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				assert.NotEmpty(s.T(), response.ImageUUID)
			}
		})
	}
}

// TestImageController_GetPublicList 测试获取公开图片列表
func (s *ControllerTestSuite) TestImageController_GetPublicList() {
	w := s.makeJSONRequest("GET", "/v1/images", nil, "")
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证响应格式
	var response api.ListImageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)
}

// TestImageController_GetUserPublicList 测试获取用户公开图片列表
func (s *ControllerTestSuite) TestImageController_GetUserPublicList() {
	url := fmt.Sprintf("/v1/images/user/%s/public?offset=1&limit=10", s.testUser.User.UserUUID)
	w := s.makeJSONRequest("GET", url, nil, "")
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证响应格式
	var response api.ListImageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)
}

// TestImageController_GetUserImagesList 测试获取用户图片列表（需要认证）
func (s *ControllerTestSuite) TestImageController_GetUserImagesList() {
	token := s.generateToken(s.testUser.User.UserUUID)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "成功获取用户图片列表",
			token:          token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "未提供token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("/v1/images/user/%s/images?offset=1&limit=10", s.testUser.User.UserUUID)
			w := s.makeJSONRequest("GET", url, nil, tt.token)
			assert.Equal(s.T(), tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response api.ListImageResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(s.T(), err)
			}
		})
	}
}

// TestChangePassword 测试修改密码
func (s *ControllerTestSuite) TestChangePassword() {
	tests := []struct {
		name           string
		request        api.ChangePasswordRequest
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "成功修改密码",
			request: api.ChangePasswordRequest{
				OldPassword: s.testUser.Password,
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "旧密码错误",
			request: api.ChangePasswordRequest{
				OldPassword: "wrongpassword",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("/v1/auth/change-password/%s", s.testUser.Email)
			w := s.makeJSONRequest("POST", url, tt.request, "")
			assert.Equal(s.T(), tt.expectedStatus, w.Code)
		})
	}
}

// TestUserRegistrationAndLogin 测试用户注册和登录流程
func (s *ControllerTestSuite) TestUserRegistrationAndLogin() {
	// 1. 用户注册
	request := api.CreateUserRequest{
		Nickname: faker.Name(),
		Email:    faker.Email(),
		Password: "password123",
	}

	w := s.makeJSONRequest("POST", "/v1/auth/register", request, "")
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 2. 用户登录
	loginRequest := api.LoginRequest{
		Email:    s.testUser.Email,
		Password: s.testUser.Password,
		SeedTime: time.Now().Unix(),
	}

	w = s.makeJSONRequest("POST", "/v1/auth/login", loginRequest, "")
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证登录成功
	var loginResponse api.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), loginResponse.Token)

	url := fmt.Sprintf("/v1/users/%s", s.testUser.Email)
	w = s.makeJSONRequest("GET", url, nil, loginResponse.Token)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var userInfo api.GetUserInfoResponse
	err = json.Unmarshal(w.Body.Bytes(), &userInfo)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testUser.Email, userInfo.Email)
	assert.Equal(s.T(), s.testUser.User.UserUUID, userInfo.UserUUID)

	// 验证token有效性
	jwtC, err := token.ParseToken(loginResponse.Token)
	require.NoError(s.T(), err)
	userUUID := jwtC.UserUUID
	assert.Equal(s.T(), userInfo.UserUUID, userUUID)
}

// TestImageUploadAndManagement 测试图片上传和管理流程
func (s *ControllerTestSuite) TestImageUploadAndManagement() {
	// 获取用户token
	token := s.generateToken(s.testUser.User.UserUUID)

	// 1. 上传图片
	testImagePath := s.createTestImageFile("integration_test.jpg")
	fields := map[string]string{
		"is_public": "true",
		"tags":      "integration,test,upload",
	}

	w := s.makeMultipartRequest("/v1/images", fields, testImagePath, token)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证上传成功
	var uploadResponse api.CreateImageResponse
	err := json.Unmarshal(w.Body.Bytes(), &uploadResponse)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), uploadResponse.ImageUUID)
	imageUUID := uploadResponse.ImageUUID

	// 2. 获取图片信息
	url := fmt.Sprintf("/v1/images/%s", imageUUID)
	w = s.makeJSONRequest("GET", url, nil, token)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证图片信息
	var imageInfo api.GetImageInfoResponse
	err = json.Unmarshal(w.Body.Bytes(), &imageInfo)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), imageUUID, imageInfo.ImageUUID)
	assert.Equal(s.T(), s.testUser.User.UserUUID, imageInfo.UserUUID)
	assert.True(s.T(), imageInfo.IsPublic)

	// 3. 更新图片标签
	updateRequest := api.UpdateImageTagsRequest{
		Tags: []string{"updated", "integration", "test"},
	}
	w = s.makeJSONRequest("PATCH", url, updateRequest, token)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 4. 验证标签更新
	w = s.makeJSONRequest("GET", url, nil, token)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &imageInfo)
	require.NoError(s.T(), err)
	assert.ElementsMatch(s.T(), updateRequest.Tags, imageInfo.Tags)

	// 5. 删除图片
	w = s.makeJSONRequest("DELETE", url, nil, token)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 6. 验证图片已删除
	w = s.makeJSONRequest("GET", url, nil, token)
	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

// TestSuite 运行测试套件
func TestControllerIntegration(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
