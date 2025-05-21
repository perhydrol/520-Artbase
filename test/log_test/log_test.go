package log_test

import (
	"demo520/internal/pkg/log"
	"testing"
)

func TestInitLogger(t *testing.T) {

	// 调用你的 log.Init 函数
	log.Init(nil)
	log.Infow("Hello World-info")
	log.Warnw("Hello World-warn")
	log.Errorw("Hello World-error")
	log.Fatalw("Hello World-fatal")
	log.Panicw("Hello World-panic")
}
