package demo520

import (
	"demo520/internal/520/store"
	"demo520/internal/pkg/config"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"demo520/pkg/db"
)

// initConfig 初始化配置
func initConfig() {
	if err := config.Init(cfgFile); err != nil {
		log.Errorw("Failed to initialize configuration", "err", err)
	}
	log.Debugw("Configuration initialized successfully")
}

// LogOptions 从配置中读取日志配置，构建 `*log.LogConfig` 并返回
func LogOptions() *log.LogConfig {
	logConfig := config.GetLog()
	return &log.LogConfig{
		DisableCaller:     logConfig.DisableCaller,
		DisableStacktrace: logConfig.DisableStacktrace,
		Level:             logConfig.Level,
		Encoding:          logConfig.Encoding,
		OutputPaths:       logConfig.OutputPaths,
	}
}

// initStore 读取 db 配置，创建 gorm.DB 实例，并初始化 demo520 store 层
func initStore() error {
	dbConfig := config.GetDB()
	dbOptions := &db.MySQLOptions{
		Host:                  dbConfig.Host,
		Username:              dbConfig.Username,
		Password:              dbConfig.Password,
		Database:              dbConfig.Database,
		MaxIdleConnections:    dbConfig.MaxIdleConnections,
		MaxOpenConnections:    dbConfig.MaxOpenConnections,
		MaxConnectionLifeTime: dbConfig.MaxConnectionLifeTime,
		LogLevel:              dbConfig.LogLevel,
	}

	ins, err := db.NewMySQL(dbOptions)
	if err != nil {
		return err
	}
	// 自动迁移
	if err := ins.AutoMigrate(&model.UserM{}); err != nil {
		return err
	}

	if err := ins.AutoMigrate(&model.NewImageM{}); err != nil {
		return err
	}

	if err := ins.AutoMigrate(&model.ImageTagM{}); err != nil {
		return err
	}

	_ = store.NewStore(ins)

	return nil
}
