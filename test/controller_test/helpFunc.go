package controller_test

import (
	"bytes"
	"demo520/internal/520/biz"
	"demo520/internal/520/biz/image"
	"demo520/internal/520/controller/user"
	"demo520/internal/520/store"
	"encoding/json"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
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

func getImageBiz(db *gorm.DB) image.ImageBiz {
	iStore := store.NewStore(db)
	biz := biz.NewIBiz(iStore)
	imageBiz := biz.Images()
	return imageBiz
}

func getUserController(db *gorm.DB) *user.UserController {
	iStore := store.NewStore(db)
	return user.NewUserController(iStore)
}
