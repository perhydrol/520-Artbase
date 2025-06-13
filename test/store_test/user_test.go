package store_test

import (
	"context"
	"demo520/internal/520/store"
	"demo520/internal/pkg/model"
	"demo520/pkg/auth"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDatabase() (*gorm.DB, error) {
	// 3. 构造 DSN
	dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3316)/testdb?charset=utf8mb4&parseTime=True&loc=Local", "testpassword")

	// 4. 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.UserM{}); err != nil {
		return nil, err
	}

	return db, nil
}

func TestUserStore(t *testing.T) {
	db, err := setupDatabase()
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	userStore := store.NewStore(db).User()
	ctx := context.Background()

	// 创建用户
	password := faker.Password()
	user := &model.UserM{
		UserUUID: uuid.New().String(),
		Password: password,
		Nickname: faker.Name(),
		Email:    faker.Email(),
	}

	err = userStore.Create(ctx, user)
	assert.NoError(t, err)

	// 获取用户
	fetchedUser, err := userStore.Get(ctx, user.Email)
	assert.NoError(t, err)
	assert.Equal(t, user.UserUUID, fetchedUser.UserUUID)
	assert.True(t, auth.VerifyPassword(password, fetchedUser.Password))

	// 更新用户
	fetchedUser.Nickname = "UpdatedName"
	err = userStore.Update(ctx, fetchedUser)
	assert.NoError(t, err)

	// 列出用户
	users, err := userStore.List(ctx, 0, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(*users), 1)

	// 删除用户
	err = userStore.Delete(ctx, user.UserUUID)
	assert.NoError(t, err)
}
