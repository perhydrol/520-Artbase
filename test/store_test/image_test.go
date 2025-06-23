package store_test

import (
	"context"
	"demo520/internal/520/store"
	"demo520/internal/pkg/model"
	"fmt"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var userCount = 3

func setupImageDatabase() (*gorm.DB, []model.UserM, error) {
	// 3. 构造 DSN
	dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3316)/testdb?charset=utf8mb4&parseTime=True&loc=Local", "testpassword")

	// 4. 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.UserM{}); err != nil {
		return nil, nil, err
	}
	if err := db.AutoMigrate(&model.NewImageM{}); err != nil {
		return nil, nil, err
	}
	if err := db.AutoMigrate(&model.ImageTagM{}); err != nil {
		return nil, nil, err
	}

	userStore := store.NewStore(db).User()
	ctx := context.Background()
	users := make([]model.UserM, userCount)

	for i := range users {
		users[i].UserUUID = uuid.New().String()
		users[i].Email = faker.Email()
		users[i].Password = faker.Password()
		users[i].Nickname = faker.Name()
	}
	for i := range users {
		if err := userStore.Create(ctx, &users[i]); err != nil {
			return nil, nil, err
		}
	}

	return db, users, nil
}
