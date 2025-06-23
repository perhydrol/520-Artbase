package testhelper

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"demo520/internal/520/store"
	"demo520/internal/pkg/model"
	"encoding/hex"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TestUser 测试用户数据结构
type TestUser struct {
	User     *model.UserM
	Email    string
	Password string // 原始密码，用于登录测试
}

// TestImage 测试图片数据结构
type TestImage struct {
	Image *model.NewImageM
	Tags  []string
}

// CreateTestUser 创建测试用户
func CreateTestUser(t *testing.T, db *gorm.DB, customUser *model.UserM) *TestUser {
	ctx := context.Background()
	userStore := store.NewStore(db).User()

	// 生成测试用户数据
	rawPassword := faker.Password()
	user := &model.UserM{
		UserUUID: uuid.New().String(),
		Email:    faker.Email(),
		Password: rawPassword,
		Nickname: faker.Name(),
	}

	// 如果提供了自定义用户数据，则覆盖默认值
	if customUser != nil {
		if customUser.UserUUID != "" {
			user.UserUUID = customUser.UserUUID
		}
		if customUser.Email != "" {
			user.Email = customUser.Email
		}
		if customUser.Password != "" {
			user.Password = customUser.Password
			rawPassword = customUser.Password
		}
		if customUser.Nickname != "" {
			user.Nickname = customUser.Nickname
		}
	}

	// 创建用户
	if err := userStore.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return &TestUser{
		User:     user,
		Email:    user.Email,
		Password: rawPassword,
	}
}

// CreateTestUsers 批量创建测试用户
func CreateTestUsers(t *testing.T, db *gorm.DB, count int) []*TestUser {
	users := make([]*TestUser, count)
	for i := 0; i < count; i++ {
		users[i] = CreateTestUser(t, db, nil)
	}
	return users
}

// 生成虚假的SHA256
func GenFakeSHA256() string {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// CreateTestImage 创建测试图片
func CreateTestImage(t *testing.T, db *gorm.DB, userUUID string, customImage *model.NewImageM) *TestImage {
	ctx := context.Background()
	imageStore := store.NewStore(db).Image()

	// 生成测试图片数据
	imageUUID := uuid.New()
	image := &model.NewImageM{
		ImageUUID: datatypes.BinUUID(imageUUID),
		Hash:      []byte(GenFakeSHA256()),
		Token:     []byte(""),
		UserUUID:  datatypes.BinUUID(uuid.MustParse(userUUID)),
		IsPublic:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 如果提供了自定义图片数据，则覆盖默认值
	if customImage != nil {
		if customImage.Hash != nil {
			image.Hash = customImage.Hash
		}
		if customImage.Token != nil {
			image.Token = customImage.Token
		}
		image.IsPublic = customImage.IsPublic
	}

	// 创建图片
	if err := imageStore.Create(ctx, image); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	return &TestImage{
		Image: image,
		Tags:  []string{},
	}
}

// CreateTestImageWithTags 创建带标签的测试图片
func CreateTestImageWithTags(t *testing.T, db *gorm.DB, userUUID string, tags []string, isPublic bool) *TestImage {
	ctx := context.Background()
	imageStore := store.NewStore(db).Image()

	// 创建图片
	customImage := &model.NewImageM{
		IsPublic: isPublic,
	}
	testImage := CreateTestImage(t, db, userUUID, customImage)

	// 添加标签
	if len(tags) > 0 {
		imageUUIDStr := uuid.UUID(testImage.Image.ImageUUID).String()
		if err := imageStore.AddTagsToImage(ctx, imageUUIDStr, tags); err != nil {
			t.Fatalf("failed to add tags to test image: %v", err)
		}
		testImage.Tags = tags
	}

	return testImage
}

// CreateTestImages 批量创建测试图片
func CreateTestImages(t *testing.T, db *gorm.DB, userUUID string, count int) []*TestImage {
	images := make([]*TestImage, count)
	for i := 0; i < count; i++ {
		images[i] = CreateTestImage(t, db, userUUID, nil)
	}
	return images
}

// CreateTestImageTag 创建测试图片标签
func CreateTestImageTag(t *testing.T, db *gorm.DB, imageUUID string, tag string) *model.ImageTagM {
	imageTag := &model.ImageTagM{
		Tag:       tag,
		ImageUUID: imageUUID,
		CreatedAt: time.Now(),
	}

	if err := db.Create(imageTag).Error; err != nil {
		t.Fatalf("failed to create test image tag: %v", err)
	}

	return imageTag
}

// GenerateTestTags 生成测试标签
func GenerateTestTags(count int) []string {
	tags := make([]string, count)
	tagOptions := []string{"nature", "city", "portrait", "landscape", "art", "photography", "travel", "food", "animal", "architecture"}

	for i := 0; i < count; i++ {
		if i < len(tagOptions) {
			tags[i] = tagOptions[i]
		} else {
			tags[i] = faker.Word()
		}
	}

	return tags
}
