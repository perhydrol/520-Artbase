package middleware

import (
	"bytes"
	"demo520/internal/pkg/known"
	"demo520/internal/pkg/log"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLogging HTTP请求日志中间件
func RequestLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 使用自定义日志格式，这里返回空字符串避免重复日志
		return ""
	})
}

// RequestTracing 请求追踪中间件，添加请求ID和详细的请求/响应日志
func RequestTracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		requestID := uuid.New().String()
		c.Set(known.XRequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)

		// 记录请求开始时间
		startTime := time.Now()

		// 读取请求体（如果存在）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 记录请求信息
		log.C(c).Infow("HTTP Request Started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"user_agent", c.Request.UserAgent(),
			"client_ip", c.ClientIP(),
			"content_type", c.Request.Header.Get("Content-Type"),
			"content_length", c.Request.ContentLength,
			"request_body_size", len(requestBody),
			"headers", sanitizeHeaders(c.Request.Header),
		)

		// 如果请求体不是文件上传，记录请求体内容
		if len(requestBody) > 0 && !isMultipartForm(c) {
			body := string(requestBody)
			if len(body) > 1000 {
				body = body[:1000] + "...(truncated)"
			}
			log.C(c).Debugw("Request Body", "body", body)
		}

		// 创建响应写入器来捕获响应
		writer := &responseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(startTime)

		// 记录响应信息
		log.C(c).Infow("HTTP Request Completed",
			"status_code", c.Writer.Status(),
			"duration", duration.String(),
			"duration_ms", duration.Milliseconds(),
			"response_size", c.Writer.Size(),
			"errors", c.Errors.String(),
		)

		// 记录响应体（仅在调试模式下）
		if writer.body.Len() > 0 {
			responseBody := writer.body.String()
			if len(responseBody) > 1000 {
				responseBody = responseBody[:1000] + "...(truncated)"
			}
			log.C(c).Debugw("Response Body", "body", responseBody)
		}

		// 如果有错误，记录错误详情
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.C(c).Errorw("Request Error",
					"error", err.Error(),
					"type", err.Type,
					"meta", err.Meta,
				)
			}
		}

		// 性能警告
		if duration > 5*time.Second {
			log.C(c).Warnw("Slow Request Detected",
				"duration", duration.String(),
				"threshold", "5s",
			)
		}
	}
}

// responseWriter 用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// sanitizeHeaders 清理敏感的请求头信息
func sanitizeHeaders(headers map[string][]string) map[string][]string {
	sanitized := make(map[string][]string)
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"x-api-key":     true,
		"x-auth-token":  true,
	}

	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if sensitiveHeaders[lowerKey] {
			sanitized[key] = []string{"[REDACTED]"}
		} else {
			sanitized[key] = values
		}
	}
	return sanitized
}

// isMultipartForm 检查是否为文件上传请求
func isMultipartForm(c *gin.Context) bool {
	contentType := c.Request.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "multipart/form-data")
}

// DatabaseOperationLog 数据库操作日志中间件（用于业务层）
func DatabaseOperationLog(operation, table string) func() {
	startTime := time.Now()
	log.Infow("Database Operation Started",
		"operation", operation,
		"table", table,
		"start_time", startTime,
	)

	return func() {
		duration := time.Since(startTime)
		log.Infow("Database Operation Completed",
			"operation", operation,
			"table", table,
			"duration", duration.String(),
			"duration_ms", duration.Milliseconds(),
		)

		// 慢查询警告
		if duration > 1*time.Second {
			log.Warnw("Slow Database Operation Detected",
				"operation", operation,
				"table", table,
				"duration", duration.String(),
				"threshold", "1s",
			)
		}
	}
}