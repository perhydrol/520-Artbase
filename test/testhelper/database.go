package testhelper

import (
	"demo520/internal/pkg/model"
	"demo520/pkg/db"
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDBConfig 测试数据库配置
type TestDBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

// DefaultTestDBConfig 返回默认的测试数据库配置
func DefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     "127.0.0.1",
		Port:     "3316",
		Username: "root",
		Password: "testpassword",
		Database: "testdb",
	}
}

// SetupTestDatabase 设置测试数据库连接并执行迁移
func SetupTestDatabase(t *testing.T, config *TestDBConfig) *gorm.DB {
	if config == nil {
		config = DefaultTestDBConfig()
	}

	// 构造数据库选项
	opts := &db.MySQLOptions{
		Host:                  fmt.Sprintf("%s:%s", config.Host, config.Port),
		Username:              config.Username,
		Password:              config.Password,
		Database:              config.Database,
		MaxIdleConnections:    10,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: 10 * time.Second,
		LogLevel:              int(logger.Info),
	}

	// 连接数据库
	db, err := db.NewMySQL(opts)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// 执行自动迁移
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

// AutoMigrate 执行所有模型的自动迁移
func AutoMigrate(db *gorm.DB) error {
	// 迁移用户表
	if err := db.AutoMigrate(&model.UserM{}); err != nil {
		return fmt.Errorf("failed to migrate UserM: %w", err)
	}

	// 迁移图片表
	if err := db.AutoMigrate(&model.NewImageM{}); err != nil {
		return fmt.Errorf("failed to migrate NewImageM: %w", err)
	}

	// 迁移图片标签表
	if err := db.AutoMigrate(&model.ImageTagM{}); err != nil {
		return fmt.Errorf("failed to migrate ImageTagM: %w", err)
	}

	return nil
}

// CleanupTestDatabase 清理测试数据库
func CleanupTestDatabase(t *testing.T, db *gorm.DB) {
	// 清理所有表的数据
	tables := []string{"image_tags", "images", "users"}
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			t.Logf("Warning: failed to clean table %s: %v", table, err)
		}
	}
}

// TruncateTestDatabase 截断测试数据库所有表
func TruncateTestDatabase(t *testing.T, db *gorm.DB) {
	// 禁用外键检查
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")

	// 截断所有表
	tables := []string{"image_tags", "images", "users"}
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table)).Error; err != nil {
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}

	// 重新启用外键检查
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")
}