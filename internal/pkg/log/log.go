package log

import (
	"context"
	"demo520/internal/pkg/known"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
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
	// 新增函数入口跟踪方法
	FuncEntry(args ...interface{}) func()
	FuncEntryWithContext(ctx context.Context, args ...interface{}) func()
	ErrorWithFunc(err error, msg string, keysAndValues ...interface{})
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

// getFuncName 获取调用函数的名称
func getFuncName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	funcName := runtime.FuncForPC(pc).Name()
	// 提取函数名（去掉包路径）
	if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
		funcName = funcName[lastSlash+1:]
	}
	if lastDot := strings.LastIndex(funcName, "."); lastDot >= 0 {
		funcName = funcName[lastDot+1:]
	}
	return funcName
}

// formatArgs 格式化函数参数
func formatArgs(args ...interface{}) []interface{} {
	if len(args) == 0 {
		return []interface{}{"args", "none"}
	}
	
	var formattedArgs []interface{}
	for i, arg := range args {
		key := fmt.Sprintf("arg%d", i)
		value := formatArgValue(arg)
		formattedArgs = append(formattedArgs, key, value)
	}
	return formattedArgs
}

// formatArgValue 格式化单个参数值
func formatArgValue(arg interface{}) interface{} {
	if arg == nil {
		return "<nil>"
	}
	
	v := reflect.ValueOf(arg)
	switch v.Kind() {
	case reflect.String:
		return fmt.Sprintf("\"%s\"", v.String())
	case reflect.Ptr:
		if v.IsNil() {
			return "<nil>"
		}
		return fmt.Sprintf("&%v", formatArgValue(v.Elem().Interface()))
	case reflect.Slice, reflect.Array:
		if v.Len() > 5 {
			return fmt.Sprintf("[%d items]", v.Len())
		}
		return fmt.Sprintf("%v", arg)
	case reflect.Map:
		if v.Len() > 5 {
			return fmt.Sprintf("map[%d items]", v.Len())
		}
		return fmt.Sprintf("%v", arg)
	case reflect.Struct:
		return fmt.Sprintf("%T{...}", arg)
	default:
		return arg
	}
}

// FuncEntry 记录函数入口信息，返回一个用于记录函数退出的函数
func FuncEntry(args ...interface{}) func() {
	return logger.FuncEntry(args...)
}

func (l *zapLogger) FuncEntry(args ...interface{}) func() {
	funcName := getFuncName(3) // skip: FuncEntry -> logger.FuncEntry -> actual caller
	startTime := time.Now()
	
	// 记录函数入口
	logArgs := []interface{}{"function", funcName, "action", "entry"}
	logArgs = append(logArgs, formatArgs(args...)...)
	l.z.Sugar().Debugw("Function entry", logArgs...)
	
	// 返回用于记录函数退出的函数
	return func() {
		duration := time.Since(startTime)
		l.z.Sugar().Debugw("Function exit", 
			"function", funcName, 
			"action", "exit", 
			"duration", duration.String())
	}
}

// FuncEntryWithContext 带上下文的函数入口记录
func FuncEntryWithContext(ctx context.Context, args ...interface{}) func() {
	return logger.FuncEntryWithContext(ctx, args...)
}

func (l *zapLogger) FuncEntryWithContext(ctx context.Context, args ...interface{}) func() {
	funcName := getFuncName(3)
	startTime := time.Now()
	
	// 使用带上下文的日志记录器
	ctxLogger := l.C(ctx)
	
	// 记录函数入口
	logArgs := []interface{}{"function", funcName, "action", "entry"}
	logArgs = append(logArgs, formatArgs(args...)...)
	ctxLogger.z.Sugar().Debugw("Function entry", logArgs...)
	
	// 返回用于记录函数退出的函数
	return func() {
		duration := time.Since(startTime)
		ctxLogger.z.Sugar().Debugw("Function exit", 
			"function", funcName, 
			"action", "exit", 
			"duration", duration.String())
	}
}

// ErrorWithFunc 记录错误信息并包含函数名
func ErrorWithFunc(err error, msg string, keysAndValues ...interface{}) {
	logger.ErrorWithFunc(err, msg, keysAndValues...)
}

func (l *zapLogger) ErrorWithFunc(err error, msg string, keysAndValues ...interface{}) {
	funcName := getFuncName(3)
	
	logArgs := []interface{}{"function", funcName, "error", err.Error()}
	logArgs = append(logArgs, keysAndValues...)
	
	l.z.Sugar().Errorw(msg, logArgs...)
}
