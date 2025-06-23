package controller_test

import (
	"bytes"
	"demo520/internal/520/controller/user"
	"demo520/internal/pkg/config"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/middleware"
	"demo520/pkg/api"
	"demo520/pkg/token"
	"demo520/test/testhelper"
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
	"github.com/stretchr/testify/suite"
)

// UserControllerTestSuite 用户controller测试套件
type UserControllerTestSuite struct {
	suite.Suite
	ts       *testhelper.TestSuite
	router   *gin.Engine
	userCtrl *user.UserController
	testUser *testhelper.TestUser
}

// SetupSuite 测试套件初始化
func (s *UserControllerTestSuite) SetupSuite() {
	// 初始化配置和日志
	config.Init("../../configs/demo520.yaml")
	log.Init(log.NewLogConfig())
	token.Init("test-secret-key")

	// 设置测试数据库
	s.ts = testhelper.NewTestSuite(s.T(), nil)

	// 初始化controller
	s.userCtrl = user.NewUserController(s.ts.Store)

	// 设置gin路由
	s.setupRouter()
}

// TearDownSuite 测试套件清理
func (s *UserControllerTestSuite) TearDownSuite() {
	s.ts.Cleanup(s.T())
}

// SetupTest 每个测试前的准备
func (s *UserControllerTestSuite) SetupTest() {
	// 清理数据库
	testhelper.CleanupTestDatabase(s.T(), s.ts.DB)
	// 创建测试用户
	s.testUser = testhelper.CreateTestUser(s.T(), s.ts.DB, nil)
}

// setupRouter 设置测试路由
func (s *UserControllerTestSuite) setupRouter() {
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
}

// makeJSONRequest 发送JSON请求的辅助函数
func (s *UserControllerTestSuite) makeJSONRequest(method, url string, payload interface{}, token string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// generateToken 生成测试用的JWT token
func (s *UserControllerTestSuite) generateToken(userUUID string) string {
	token, err := token.GenerateToken(userUUID)
	require.NoError(s.T(), err)
	return token
}

// TestCreate_Success 测试成功创建用户
func (s *UserControllerTestSuite) TestCreate_Success() {
	request := api.CreateUserRequest{
		Nickname: faker.Name(),
		Email:    faker.Email(),
		Password: "password123",
	}

	w := s.makeJSONRequest("POST", "/v1/auth/register", request, "")

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证响应
	var response api.UserInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), request.Email, response.Email)
	assert.Equal(s.T(), request.Nickname, response.Nickname)
	assert.NotEmpty(s.T(), response.UserUUID)
	assert.NotEmpty(s.T(), response.CreateAt)
}

// TestCreate_ValidationErrors 测试创建用户时的验证错误
func (s *UserControllerTestSuite) TestCreate_ValidationErrors() {
	tests := []struct {
		name    string
		request api.CreateUserRequest
		desc    string
	}{
		{
			name: "邮箱格式错误",
			request: api.CreateUserRequest{
				Nickname: faker.Name(),
				Email:    "invalid-email",
				Password: "password123",
			},
			desc: "应该返回400错误当邮箱格式不正确时",
		},
		{
			name: "密码太短",
			request: api.CreateUserRequest{
				Nickname: faker.Name(),
				Email:    faker.Email(),
				Password: "123",
			},
			desc: "应该返回400错误当密码少于6个字符时",
		},
		{
			name: "昵称太短",
			request: api.CreateUserRequest{
				Nickname: "abc",
				Email:    faker.Email(),
				Password: "password123",
			},
			desc: "应该返回400错误当昵称少于6个字符时",
		},
		{
			name: "昵称太长",
			request: api.CreateUserRequest{
				Nickname: "this_is_a_very_long_nickname_that_exceeds_the_maximum_allowed_length_of_64_characters",
				Email:    faker.Email(),
				Password: "password123",
			},
			desc: "应该返回400错误当昵称超过64个字符时",
		},
		{
			name: "缺少必填字段",
			request: api.CreateUserRequest{
				Nickname: "",
				Email:    "",
				Password: "",
			},
			desc: "应该返回400错误当缺少必填字段时",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := s.makeJSONRequest("POST", "/v1/auth/register", tt.request, "")
			assert.Equal(s.T(), http.StatusBadRequest, w.Code, tt.desc)
		})
	}
}

// TestCreate_DuplicateEmail 测试重复邮箱注册
func (s *UserControllerTestSuite) TestCreate_DuplicateEmail() {
	// 使用已存在的用户邮箱
	request := api.CreateUserRequest{
		Nickname: faker.Name(),
		Email:    s.testUser.Email,
		Password: "password123",
	}

	w := s.makeJSONRequest("POST", "/v1/auth/register", request, "")

	// 应该返回冲突错误
	assert.Equal(s.T(), http.StatusConflict, w.Code)
}

// TestLogin_Success 测试成功登录
func (s *UserControllerTestSuite) TestLogin_Success() {
	request := api.LoginRequest{
		Email:    s.testUser.User.Email,
		Password: s.testUser.Password,
		SeedTime: time.Now().Unix(),
	}

	w := s.makeJSONRequest("POST", "/v1/auth/login", request, "")

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证响应
	var response api.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), response.Token)

	// 验证token有效性
	userUUID, err := token.ParseToken(response.Token)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testUser.User.UserUUID, userUUID.UserUUID)
}

// TestLogin_InvalidCredentials 测试无效凭据
func (s *UserControllerTestSuite) TestLogin_InvalidCredentials() {
	tests := []struct {
		name    string
		request api.LoginRequest
		desc    string
	}{
		{
			name: "错误密码",
			request: api.LoginRequest{
				Email:    s.testUser.User.Email,
				Password: "wrongpassword",
				SeedTime: time.Now().Unix(),
			},
			desc: "应该返回401错误当密码错误时",
		},
		{
			name: "用户不存在",
			request: api.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
				SeedTime: time.Now().Unix(),
			},
			desc: "应该返回401错误当用户不存在时",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := s.makeJSONRequest("POST", "/v1/auth/login", tt.request, "")
			assert.Equal(s.T(), http.StatusUnauthorized, w.Code, tt.desc)
		})
	}
}

// TestLogin_ValidationErrors 测试登录时的验证错误
func (s *UserControllerTestSuite) TestLogin_ValidationErrors() {
	tests := []struct {
		name    string
		request api.LoginRequest
		desc    string
	}{
		{
			name: "邮箱格式错误",
			request: api.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
				SeedTime: time.Now().Unix(),
			},
			desc: "应该返回400错误当邮箱格式不正确时",
		},
		{
			name: "密码太短",
			request: api.LoginRequest{
				Email:    faker.Email(),
				Password: "123",
				SeedTime: time.Now().Unix(),
			},
			desc: "应该返回400错误当密码少于6个字符时",
		},
		{
			name: "缺少SeedTime",
			request: api.LoginRequest{
				Email:    faker.Email(),
				Password: "password123",
				SeedTime: 0,
			},
			desc: "应该返回400错误当缺少SeedTime时",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := s.makeJSONRequest("POST", "/v1/auth/login", tt.request, "")
			assert.Equal(s.T(), http.StatusBadRequest, w.Code, tt.desc)
		})
	}
}

// TestGet_Success 测试成功获取用户信息
func (s *UserControllerTestSuite) TestGet_Success() {
	token := s.generateToken(s.testUser.User.UserUUID)
	url := fmt.Sprintf("/v1/users/%s", s.testUser.Email)

	w := s.makeJSONRequest("GET", url, nil, token)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证响应内容
	var response api.GetUserInfoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testUser.Email, response.Email)
	assert.Equal(s.T(), s.testUser.User.UserUUID, response.UserUUID)
	assert.Equal(s.T(), s.testUser.User.Nickname, response.Nickname)
	assert.NotEmpty(s.T(), response.CreateAt)
}

// TestGet_Unauthorized 测试未授权访问
func (s *UserControllerTestSuite) TestGet_Unauthorized() {
	tests := []struct {
		name  string
		token string
		desc  string
	}{
		{
			name:  "无效token",
			token: "invalid-token",
			desc:  "应该返回401错误当token无效时",
		},
		{
			name:  "过期token",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			desc:  "应该返回401错误当token过期时",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("/v1/users/%s", s.testUser.Email)
			w := s.makeJSONRequest("GET", url, nil, tt.token)
			assert.Equal(s.T(), http.StatusUnauthorized, w.Code, tt.desc)
		})
	}
}

// TestGet_UserNotFound 测试用户不存在
func (s *UserControllerTestSuite) TestGet_UserNotFound() {
	token := s.generateToken(s.testUser.User.UserUUID)
	url := "/v1/users/nonexistent@example.com"

	w := s.makeJSONRequest("GET", url, nil, token)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

// TestUpdate_Success 测试成功更新用户信息
func (s *UserControllerTestSuite) TestUpdate_Success() {
	token := s.generateToken(s.testUser.User.UserUUID)
	request := api.UpdateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
	}
	url := fmt.Sprintf("/v1/users/%s", s.testUser.Email)

	w := s.makeJSONRequest("PATCH", url, request, token)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证用户信息已更新
	updatedUser, err := s.ts.Store.User().Get(s.ts.Ctx, request.Email)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), request.Email, updatedUser.Email)
	assert.Equal(s.T(), request.Nickname, updatedUser.Nickname)
}

// TestUpdate_ValidationErrors 测试更新用户时的验证错误
func (s *UserControllerTestSuite) TestUpdate_ValidationErrors() {
	token := s.generateToken(s.testUser.User.UserUUID)

	tests := []struct {
		name    string
		request api.UpdateUserRequest
		desc    string
	}{
		{
			name: "邮箱格式错误",
			request: api.UpdateUserRequest{
				Email:    "invalid-email",
				Nickname: faker.Name(),
			},
			desc: "应该返回400错误当邮箱格式不正确时",
		},
		{
			name: "昵称太短",
			request: api.UpdateUserRequest{
				Email:    faker.Email(),
				Nickname: "abc",
			},
			desc: "应该返回400错误当昵称少于6个字符时",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("/v1/users/%s", s.testUser.Email)
			w := s.makeJSONRequest("PATCH", url, tt.request, token)
			assert.Equal(s.T(), http.StatusBadRequest, w.Code, tt.desc)
		})
	}
}

// TestChangePassword_Success 测试成功修改密码
func (s *UserControllerTestSuite) TestChangePassword_Success() {
	request := api.ChangePasswordRequest{
		OldPassword: s.testUser.Password,
		NewPassword: "newpassword123",
	}
	url := fmt.Sprintf("/v1/auth/change-password/%s", s.testUser.Email)

	w := s.makeJSONRequest("POST", url, request, "")

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证可以用新密码登录
	loginRequest := api.LoginRequest{
		Email:    s.testUser.Email,
		Password: request.NewPassword,
		SeedTime: time.Now().Unix(),
	}
	w = s.makeJSONRequest("POST", "/v1/auth/login", loginRequest, "")
	assert.Equal(s.T(), http.StatusOK, w.Code)
}

// TestChangePassword_InvalidOldPassword 测试旧密码错误
func (s *UserControllerTestSuite) TestChangePassword_InvalidOldPassword() {
	request := api.ChangePasswordRequest{
		OldPassword: "wrongpassword",
		NewPassword: "newpassword123",
	}
	url := fmt.Sprintf("/v1/auth/change-password/%s", s.testUser.Email)

	w := s.makeJSONRequest("POST", url, request, "")

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

// TestChangePassword_ValidationErrors 测试修改密码时的验证错误
func (s *UserControllerTestSuite) TestChangePassword_ValidationErrors() {
	tests := []struct {
		name    string
		request api.ChangePasswordRequest
		desc    string
	}{
		{
			name: "新密码太短",
			request: api.ChangePasswordRequest{
				OldPassword: s.testUser.Password,
				NewPassword: "123",
			},
			desc: "应该返回400错误当新密码少于6个字符时",
		},
		{
			name: "旧密码太短",
			request: api.ChangePasswordRequest{
				OldPassword: "123",
				NewPassword: "newpassword123",
			},
			desc: "应该返回400错误当旧密码少于6个字符时",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("/v1/auth/change-password/%s", s.testUser.Email)
			w := s.makeJSONRequest("POST", url, tt.request, "")
			assert.Equal(s.T(), http.StatusBadRequest, w.Code, tt.desc)
		})
	}
}

// TestUserController 运行用户controller测试套件
func TestUserController(t *testing.T) {
	suite.Run(t, new(UserControllerTestSuite))
}
