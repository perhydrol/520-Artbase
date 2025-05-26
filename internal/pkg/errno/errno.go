package errno

import (
	"errors"
	"fmt"
)

type Errno struct {
	HTTP    int    `json:"http"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (err *Errno) Error() string {
	return err.Message
}

// SetMessage 设置 Errno 类型错误中的 Message 字段.
func (err *Errno) SetMessage(format string, args ...interface{}) *Errno {
	err.Message = fmt.Sprintf(format, args...)
	return err
}

// Decode 尝试从 err 中解析出业务错误码和错误信息.
func Decode(err error) (int, string, string) {
	if err == nil {
		return OK.HTTP, OK.Code, OK.Message
	}

	var typed *Errno
	switch {
	case errors.As(err, &typed):
		return typed.HTTP, typed.Code, typed.Message
	default:
	}

	// 默认返回未知错误码和错误信息. 该错误代表服务端出错
	return InternalServerError.HTTP, InternalServerError.Code, err.Error()
}
