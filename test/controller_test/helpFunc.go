package controller_test

import (
	"bytes"
	"demo520/internal/520/controller/image"
	"demo520/internal/520/controller/user"
	"demo520/internal/520/store"
	"demo520/pkg/api"
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
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// 创建测试用的 Gin Context
func createTestContext(method string, url string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		c.Request = httptest.NewRequest(method, url, bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
	} else {
		c.Request = httptest.NewRequest(method, url, nil)
	}

	return c, w
}

func genLoginReq(email, password string) *api.LoginRequest {
	return &api.LoginRequest{
		Email:    email,
		Password: password,
		SeedTime: time.Now().Unix(),
	}
}

func getImageController(db *gorm.DB) *image.ImageController {
	iStore := store.NewStore(db)
	return image.NewUserController(iStore)
}

func getUserController(db *gorm.DB) *user.UserController {
	iStore := store.NewStore(db)
	return user.NewUserController(iStore)
}

func appendJWTHeader(c *gin.Context, token string) {
	c.Request.Header.Set("Authorization", "Bearer "+token)
}

func loginAndGetToken(db *gorm.DB, email, password string) (string, error) {
	loginReq := genLoginReq(email, password)
	c, w := createTestContext("POST", "/login", &loginReq)
	getUserController(db).Login(c)
	if w.Code != http.StatusOK {
		return "", fmt.Errorf("failed to login: %v", w.Code)
	}
	var loginResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		return "", err
	}
	return loginResp["token"], nil
}

func prepareContextWithFile(t *testing.T, filePath string, createReq *api.CreateImageRequest) (*gin.Context, *httptest.ResponseRecorder) {
	// 打开本地文件
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	// 用 multipart.Writer 构造 form-data body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", filepath.Base(filePath))
	require.NoError(t, err)
	_, err = io.Copy(part, file)
	require.NoError(t, err)

	// 如果需要传递 json 字段
	jsonData, err := json.Marshal(createReq)
	require.NoError(t, err)
	err = writer.WriteField("json", string(jsonData))
	writer.Close()
	require.NoError(t, err)

	// 创建模拟请求
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 创建 gin.Context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
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
