package core

import (
	"demo520/internal/pkg/errno"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrResponse 定义了发生错误时的返回消息.
type ErrResponse struct {
	// Code 指定了业务错误码.
	Code string `json:"code"`

	// Message 包含了可以直接对外展示的错误信息.
	Message string `json:"message"`
}

func WriteResponse(c *gin.Context, err error, data interface{}) {
	if err != nil {
		hcode, code, message := errno.Decode(err)
		c.JSON(hcode, ErrResponse{Code: code, Message: message})
		return
	}
	c.JSON(http.StatusOK, data)
}

func ResponseFile(c *gin.Context, err error, filePath string) {
	if err != nil {
		hcode, code, message := errno.Decode(err)
		c.JSON(hcode, ErrResponse{Code: code, Message: message})
		return
	}
	c.File(filePath)
}
