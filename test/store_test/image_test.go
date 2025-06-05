package store

import (
	"context"
	"crypto/sha256"
	"demo520/internal/520/store"
	"demo520/internal/pkg/model"
	"fmt"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"strings"
	"testing"
)

func setupDatabase() (*gorm.DB, []model.UserM, error) {
	// 3. 构造 DSN
	dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local", "testpassword")

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
	if err := db.AutoMigrate(&model.ImageM{}); err != nil {
		return nil, nil, err
	}
	if err := db.AutoMigrate(&model.ImageTagM{}); err != nil {
		return nil, nil, err
	}

	userStore := store.NewStore(db).User()
	ctx := context.Background()
	users := make([]model.UserM, 2)

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

func TestImageStore(t *testing.T) {
	db, users, err := setupDatabase()
	if err != nil {
		t.Fatal(err)
	}

	imageStore := store.NewStore(db).Image()
	ctx := context.Background()

	images := [2]model.ImageM{}
	for i := range images {
		images[i].ImageUUID = faker.UUIDHyphenated()
		images[i].OwnerUUID = users[i].UserUUID
		images[i].Token = faker.UUIDHyphenated()
		hasher := sha256.New()
		if _, err := io.Copy(hasher, strings.NewReader(images[i].ImageUUID)); err != nil {
			t.Fatal(err)
		}
		images[i].Hash = fmt.Sprintf("%x", hasher.Sum(nil))
		images[i].IsPublic = false
		images[i].Tags = []model.ImageTagM{
			{Tag: faker.Word(), ImageUUID: images[i].ImageUUID},
			{Tag: faker.Word(), ImageUUID: images[i].ImageUUID},
			{Tag: faker.Word(), ImageUUID: images[i].ImageUUID},
		}
	}
	for i := range images {
		if err := imageStore.Create(ctx, &images[i]); err != nil {
			t.Fatal(err)
		}
	}
}
