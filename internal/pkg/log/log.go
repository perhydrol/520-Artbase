package log

import (
	"context"
	"demo520/internal/pkg/known"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

type Logger interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	Sync()
}

type zapLogger struct {
	z *zap.Logger
}

var _ Logger = &zapLogger{}

var (
	once   sync.Once
	logger *zapLogger
)

func Init(opts *LogConfig) {
	once.Do(func() {
		logger = newLogger(opts)
	})
}

func newLogger(opts *LogConfig) *zapLogger {
	if opts == nil {
		opts = NewLogConfig()
	}
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.MessageKey = "message"
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	cfg := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		DisableCaller:     opts.DisableCaller,
		DisableStacktrace: opts.DisableStacktrace,
		Encoding:          opts.Encoding,
		EncoderConfig:     encoderConfig,
		OutputPaths:       opts.OutputPaths,
		ErrorOutputPaths:  []string{"stderr"},
	}

	z, err := cfg.Build(zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		panic(err)
	}
	logger = &zapLogger{z}
	zap.RedirectStdLog(z)
	return logger
}

// Sync 调用底层 zap.Logger 的 Sync 方法，将缓存中的日志刷新到磁盘文件中. 主程序需要在退出前调用 Sync.
func Sync() { logger.Sync() }

func (l *zapLogger) Sync() {
	_ = l.z.Sync()
}

// Debugw 输出 debug 级别的日志.
func Debugw(msg string, keysAndValues ...interface{}) {
	logger.z.Sugar().Debugw(msg, keysAndValues...)
}

func (l *zapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Debugw(msg, keysAndValues...)
}

// Infow 输出 info 级别的日志.
func Infow(msg string, keysAndValues ...interface{}) {
	logger.z.Sugar().Infow(msg, keysAndValues...)
}

func (l *zapLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Infow(msg, keysAndValues...)
}

// Warnw 输出 warning 级别的日志.
func Warnw(msg string, keysAndValues ...interface{}) {
	logger.z.Sugar().Warnw(msg, keysAndValues...)
}

func (l *zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Warnw(msg, keysAndValues...)
}

// Errorw 输出 error 级别的日志.
func Errorw(msg string, keysAndValues ...interface{}) {
	logger.z.Sugar().Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Errorw(msg, keysAndValues...)
}

// Panicw 输出 panic 级别的日志.
func Panicw(msg string, keysAndValues ...interface{}) {
	logger.z.Sugar().Panicw(msg, keysAndValues...)
}

func (l *zapLogger) Panicw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Panicw(msg, keysAndValues...)
}

// Fatalw 输出 fatal 级别的日志.
func Fatalw(msg string, keysAndValues ...interface{}) {
	logger.z.Sugar().Fatalw(msg, keysAndValues...)
}

func (l *zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Fatalw(msg, keysAndValues...)
}

// C 解析传入的 context，尝试提取关注的键值，并添加到 zap.Logger 结构化日志中.
func C(ctx context.Context) *zapLogger {
	return logger.C(ctx)
}

func (l *zapLogger) C(ctx context.Context) *zapLogger {
	lc := l.clone()

	if requestID := ctx.Value(known.XRequestIDKey); requestID != nil {
		lc.z = lc.z.With(zap.Any(known.XRequestIDKey, requestID))
	}

	if userID := ctx.Value(known.XUsernameKey); userID != nil {
		lc.z = lc.z.With(zap.Any(known.XUsernameKey, userID))
	}

	return lc
}

// clone 深度拷贝 zapLogger.
func (l *zapLogger) clone() *zapLogger {
	lc := *l
	return &lc
}
