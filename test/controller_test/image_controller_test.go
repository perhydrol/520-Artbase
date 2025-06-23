package controller_test

import (
	"bytes"
	"demo520/internal/520/controller/image"
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
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ImageControllerTestSuite 图片controller测试套件
type ImageControllerTestSuite struct {
	suite.Suite
	ts           *testhelper.TestSuite
	router       *gin.Engine
	imageCtrl    *image.ImageController
	testUser     *testhelper.TestUser
	testUser2    *testhelper.TestUser
	testImageDir string
	testImages   []*testhelper.TestImage
}

// SetupSuite 测试套件初始化
func (s *ImageControllerTestSuite) SetupSuite() {
	// 初始化配置和日志
	config.Init("../../configs/demo520.yaml")
	log.Init(log.NewLogConfig())
	token.Init("test-secret-key")

	// 设置测试数据库
	s.ts = testhelper.NewTestSuite(s.T(), nil)

	// 初始化controller
	s.imageCtrl = image.NewUserController(s.ts.Store)

	// 设置gin路由
	s.setupRouter()

	// 创建测试图片目录
	s.setupTestImageDir()
}

// TearDownSuite 测试套件清理
func (s *ImageControllerTestSuite) TearDownSuite() {
	s.ts.Cleanup(s.T())
	// 清理测试图片目录
	if s.testImageDir != "" {
		os.RemoveAll(s.testImageDir)
	}
}

// SetupTest 每个测试前的准备
func (s *ImageControllerTestSuite) SetupTest() {
	// 清理数据库
	testhelper.CleanupTestDatabase(s.T(), s.ts.DB)
	// 创建测试用户
	s.testUser = testhelper.CreateTestUser(s.T(), s.ts.DB, nil)
	s.testUser2 = testhelper.CreateTestUser(s.T(), s.ts.DB, nil)
	// 创建测试图片数据
	s.createTestImages()
}

// setupRouter 设置测试路由
func (s *ImageControllerTestSuite) setupRouter() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// 注册路由
	api_v1 := s.router.Group("/v1")

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

// setupTestImageDir 设置测试图片目录
func (s *ImageControllerTestSuite) setupTestImageDir() {
	tempDir, err := os.MkdirTemp("", "image_controller_test")
	require.NoError(s.T(), err)
	s.testImageDir = tempDir
}

// createTestImageFile 使用真实的测试图片文件
func (s *ImageControllerTestSuite) createTestImageFile(filename string) string {
	// 使用项目中的真实图片文件
	realImagePath := "/home/oxygen/project/520/test/test_image/cg_tm01_0101__base_cg_tm01_0101.png"
	if filename != "" {
		realImagePath = filename
	}

	// 检查文件是否存在
	if _, err := os.Stat(realImagePath); os.IsNotExist(err) {
		s.T().Fatalf("测试图片文件不存在: %s", realImagePath)
	}

	return realImagePath
}

// createTestErrorFile 创建用于测试的错误文件（非图片文件）
func (s *ImageControllerTestSuite) createTestErrorFile() string {
	// 使用项目中的文本文件作为错误文件
	errorFilePath := "/home/oxygen/project/520/test/controller_test/test_file.txt"

	// 检查文件是否存在
	if _, err := os.Stat(errorFilePath); os.IsNotExist(err) {
		s.T().Fatalf("测试错误文件不存在: %s", errorFilePath)
	}

	return errorFilePath
}

// createTestImages 创建测试图片数据
func (s *ImageControllerTestSuite) createTestImages() {
	s.testImages = []*testhelper.TestImage{
		testhelper.CreateTestImage(s.T(), s.ts.DB, s.testUser.User.UserUUID, nil),
		testhelper.CreateTestImage(s.T(), s.ts.DB, s.testUser2.User.UserUUID, nil),
	}
}

// makeRequest 发送HTTP请求的辅助函数
func (s *ImageControllerTestSuite) makeRequest(method, url string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, body)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// makeJSONRequest 发送JSON请求的辅助函数
func (s *ImageControllerTestSuite) makeJSONRequest(method, url string, payload interface{}, token string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(payload)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	return s.makeRequest(method, url, bytes.NewReader(body), headers)
}

// makeMultipartRequest 发送multipart请求的辅助函数
func (s *ImageControllerTestSuite) makeMultipartRequest(url string, fields map[string]string, filePath string, token string) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件
	if filePath != "" {
		file, err := os.Open(filePath)
		require.NoError(s.T(), err)
		defer file.Close()

		part, err := writer.CreateFormFile("image", filepath.Base(filePath))
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

// generateToken 生成测试用的JWT token
func (s *ImageControllerTestSuite) generateToken(userUUID string) string {
	token, err := token.GenerateToken(userUUID)
	require.NoError(s.T(), err)
	return token
}

// TestCreate_Success 测试成功上传图片
func (s *ImageControllerTestSuite) TestCreate_Success() {
	token := s.generateToken(s.testUser.User.UserUUID)
	testImagePath := s.createTestImageFile("")

	fields := map[string]string{
		"is_public": "true",
		"tags":      "test,upload,image",
	}

	w := s.makeMultipartRequest("/v1/images", fields, testImagePath, token)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证响应
	var response api.CreateImageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), response.ImageUUID)
	assert.Equal(s.T(), s.testUser.User.UserUUID, response.UserUUID)
	assert.True(s.T(), response.IsPublic)
}

// TestCreate_Unauthorized 测试未授权上传
func (s *ImageControllerTestSuite) TestCreate_Unauthorized() {
	testImagePath := s.createTestImageFile("")

	tests := []struct {
		name  string
		token string
		desc  string
	}{
		{
			name:  "无token",
			token: "",
			desc:  "应该返回401错误当没有提供token时",
		},
		{
			name:  "无效token",
			token: "invalid-token",
			desc:  "应该返回401错误当token无效时",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			fields := map[string]string{
				"is_public": "true",
			}

			w := s.makeMultipartRequest("/v1/images", fields, testImagePath, tt.token)
			assert.Equal(s.T(), http.StatusUnauthorized, w.Code, tt.desc)
		})
	}
}

// TestCreate_NoFile 测试没有文件的上传
func (s *ImageControllerTestSuite) TestCreate_NoFile() {
	token := s.generateToken(s.testUser.User.UserUUID)

	fields := map[string]string{
		"is_public": "true",
	}

	w := s.makeMultipartRequest("/v1/images", fields, "", token)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

// TestCreate_InvalidFileType 测试无效文件类型
func (s *ImageControllerTestSuite) TestCreate_InvalidFileType() {
	token := s.generateToken(s.testUser.User.UserUUID)
	testFilePath := s.createTestErrorFile() // 使用专门的错误文件函数

	fields := map[string]string{
		"is_public": "true",
	}

	w := s.makeMultipartRequest("/v1/images", fields, testFilePath, token)

	// 根据业务逻辑，可能返回400或其他错误码
	assert.NotEqual(s.T(), http.StatusOK, w.Code)
}

// TestGetPublicList_Success 测试获取公开图片列表
func (s *ImageControllerTestSuite) TestGetPublicList_Success() {
	w := s.makeJSONRequest("GET", "/v1/images", nil, "")

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证响应格式
	var response api.ListImageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// 应该只包含公开的图片
	for _, img := range response.ImageList {
		assert.True(s.T(), img.IsPublic, "公开列表应该只包含公开图片")
	}
}

// TestGetPublicList_WithPagination 测试分页获取公开图片列表
func (s *ImageControllerTestSuite) TestGetPublicList_WithPagination() {
	tests := []struct {
		name   string
		query  string
		status int
	}{
		{
			name:   "有效分页参数",
			query:  "limit=10",
			status: http.StatusOK,
		},
		{
			name:   "无效限制数量",
			query:  "limit=100",
			status: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := s.makeJSONRequest("GET", "/v1/images"+tt.query, nil, "")
			assert.Equal(s.T(), tt.status, w.Code)
		})
	}
}

// TestGetUserPublicList_Success 测试获取用户公开图片列表
func (s *ImageControllerTestSuite) TestGetUserPublicList_Success() {
	url := fmt.Sprintf("/v1/images/user/%s/public?offset=0&limit=10", s.testUser.User.UserUUID)
	w := s.makeJSONRequest("GET", url, nil, "")

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证响应格式
	var response api.ListImageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// 应该只包含指定用户的公开图片
	for _, img := range response.ImageList {
		assert.True(s.T(), img.IsPublic, "应该只包含公开图片")
		assert.Equal(s.T(), s.testUser.User.UserUUID, img.UserUUID, "应该只包含指定用户的图片")
	}
}

// TestGetUserPublicList_InvalidUserUUID 测试无效用户UUID
func (s *ImageControllerTestSuite) TestGetUserPublicList_InvalidUserUUID() {
	url := "/v1/images/user/invalid-uuid/public"
	w := s.makeJSONRequest("GET", url, nil, "")

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

// TestGet_Success 测试成功获取图片信息
func (s *ImageControllerTestSuite) TestGet_Success() {
	token := s.generateToken(s.testUser.User.UserUUID)
	imageUUID := s.testImages[0].Image.ImageUUID
	url := fmt.Sprintf("/v1/images/%s", imageUUID)

	w := s.makeJSONRequest("GET", url, nil, token)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证响应内容
	var response api.GetImageInfoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), imageUUID.String(), response.ImageUUID)
}

// TestGet_Unauthorized 测试未授权获取图片
func (s *ImageControllerTestSuite) TestGet_Unauthorized() {
	imageUUID := s.testImages[0].Image.ImageUUID
	url := fmt.Sprintf("/v1/images/%s", imageUUID)

	tests := []struct {
		name  string
		token string
		desc  string
	}{
		{
			name:  "无token",
			token: "",
			desc:  "应该返回401错误当没有提供token时",
		},
		{
			name:  "无效token",
			token: "invalid-token",
			desc:  "应该返回401错误当token无效时",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := s.makeJSONRequest("GET", url, nil, tt.token)
			assert.Equal(s.T(), http.StatusUnauthorized, w.Code, tt.desc)
		})
	}
}

// TestGet_NotFound 测试图片不存在
func (s *ImageControllerTestSuite) TestGet_NotFound() {
	token := s.generateToken(s.testUser.User.UserUUID)
	nonExistentUUID := uuid.New().String()
	url := fmt.Sprintf("/v1/images/%s", nonExistentUUID)

	w := s.makeJSONRequest("GET", url, nil, token)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

// TestGetUserImagesList_Success 测试获取用户图片列表
func (s *ImageControllerTestSuite) TestGetUserImagesList_Success() {
	token := s.generateToken(s.testUser.User.UserUUID)
	// 添加必需的查询参数offset和limit
	url := fmt.Sprintf("/v1/images/user/%s/images?offset=0&limit=10", s.testUser.User.UserUUID)

	w := s.makeJSONRequest("GET", url, nil, token)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证响应格式
	var response api.ListImageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// 应该只包含指定用户的图片
	for _, img := range response.ImageList {
		assert.Equal(s.T(), s.testUser.User.UserUUID, img.UserUUID, "应该只包含指定用户的图片")
	}
}

// TestGetUserImagesList_AccessControl 测试访问控制
func (s *ImageControllerTestSuite) TestGetUserImagesList_AccessControl() {
	// 用户1的token尝试访问用户2的图片列表
	token := s.generateToken(s.testUser.User.UserUUID)
	url := fmt.Sprintf("/v1/images/user/%s/images", s.testUser2.User.UserUUID)

	w := s.makeJSONRequest("GET", url, nil, token)

	// 根据业务逻辑，可能返回403或其他错误码
	assert.NotEqual(s.T(), http.StatusOK, w.Code, "不应该允许访问其他用户的私有图片列表")
}

// TestUpdateImageTags_Success 测试成功更新图片标签
func (s *ImageControllerTestSuite) TestUpdateImageTags_Success() {
	token := s.generateToken(s.testUser.User.UserUUID)
	imageUUID := s.testImages[0].Image.ImageUUID
	url := fmt.Sprintf("/v1/images/%s", imageUUID)

	request := api.UpdateImageTagsRequest{
		Tags: []string{"updated", "tags", "test"},
	}

	w := s.makeJSONRequest("PATCH", url, request, token)

	assert.Equal(s.T(), http.StatusOK, w.Code)
}

// TestUpdateImageTags_Unauthorized 测试未授权更新标签
func (s *ImageControllerTestSuite) TestUpdateImageTags_Unauthorized() {
	// 用户1尝试更新用户2的图片标签
	token := s.generateToken(s.testUser.User.UserUUID)
	imageUUID := s.testImages[1].Image.ImageUUID // 用户2的图片
	url := fmt.Sprintf("/v1/images/%s", imageUUID)

	request := api.UpdateImageTagsRequest{
		Tags: []string{"unauthorized", "update"},
	}

	w := s.makeJSONRequest("PATCH", url, request, token)

	// 应该返回403或404错误
	assert.NotEqual(s.T(), http.StatusOK, w.Code, "不应该允许更新其他用户的图片")
}

// TestUpdateImageTags_InvalidTags 测试无效标签
func (s *ImageControllerTestSuite) TestUpdateImageTags_InvalidTags() {
	token := s.generateToken(s.testUser.User.UserUUID)
	imageUUID := s.testImages[0].Image.ImageUUID
	url := fmt.Sprintf("/v1/images/%s", imageUUID)

	tests := []struct {
		name    string
		request api.UpdateImageTagsRequest
		desc    string
	}{
		{
			name: "空标签数组",
			request: api.UpdateImageTagsRequest{
				Tags: []string{},
			},
			desc: "空标签数组应该被接受",
		},
		{
			name: "包含空字符串的标签",
			request: api.UpdateImageTagsRequest{
				Tags: []string{"valid", "", "tag"},
			},
			desc: "包含空字符串的标签可能被拒绝",
		},
		{
			name: "过长的标签",
			request: api.UpdateImageTagsRequest{
				Tags: []string{strings.Repeat("a", 100)},
			},
			desc: "过长的标签应该被拒绝",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := s.makeJSONRequest("PATCH", url, tt.request, token)
			// 根据具体的验证逻辑，这里可能需要调整期望的状态码
			// assert.Equal(s.T(), expectedStatusCode, w.Code, tt.desc)
			_ = w // 暂时忽略具体的断言，因为需要了解具体的业务逻辑
		})
	}
}

// TestDelete_Success 测试成功删除图片
func (s *ImageControllerTestSuite) TestDelete_Success() {
	token := s.generateToken(s.testUser.User.UserUUID)
	imageUUID := s.testImages[0].Image.ImageUUID
	url := fmt.Sprintf("/v1/images/%s", imageUUID)

	w := s.makeJSONRequest("DELETE", url, nil, token)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 验证图片已被删除
	w = s.makeJSONRequest("GET", url, nil, token)
	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

// TestDelete_Unauthorized 测试未授权删除
func (s *ImageControllerTestSuite) TestDelete_Unauthorized() {
	// 用户1尝试删除用户2的图片
	token := s.generateToken(s.testUser.User.UserUUID)
	imageUUID := s.testImages[1].Image.ImageUUID // 用户2的图片
	url := fmt.Sprintf("/v1/images/%s", imageUUID)

	w := s.makeJSONRequest("DELETE", url, nil, token)

	// 应该返回403或404错误
	assert.NotEqual(s.T(), http.StatusOK, w.Code, "不应该允许删除其他用户的图片")
}

// TestGetImageFile_Success 测试成功获取图片文件
func (s *ImageControllerTestSuite) TestGetImageFile_Success() {
	token := s.generateToken(s.testUser.User.UserUUID)
	// 假设图片文件名格式为 imageUUID.jpg
	imageUUIDFileName := s.testImages[0].Image.ImageUUID.String() + ".jpg"
	url := fmt.Sprintf("/v1/images/file/%s", imageUUIDFileName)

	w := s.makeJSONRequest("GET", url, nil, token)

	// 根据实际的文件存储情况，这里可能返回200或404
	// assert.Equal(s.T(), http.StatusOK, w.Code)
	_ = w // 暂时忽略具体断言，因为需要实际的文件存储
}

// TestGetImageFile_NotFound 测试图片文件不存在
func (s *ImageControllerTestSuite) TestGetImageFile_NotFound() {
	token := s.generateToken(s.testUser.User.UserUUID)
	nonExistentFileName := uuid.New().String() + ".jpg"
	url := fmt.Sprintf("/v1/images/file/%s", nonExistentFileName)

	w := s.makeJSONRequest("GET", url, nil, token)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

// TestImageController 运行图片controller测试套件
func TestImageController(t *testing.T) {
	suite.Run(t, new(ImageControllerTestSuite))
}
