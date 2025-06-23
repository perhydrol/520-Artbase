package testhelper

import (
	"context"
	"demo520/internal/520/store"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestSuite 测试套件结构
type TestSuite struct {
	DB    *gorm.DB
	Store store.IStore
	Ctx   context.Context
}

// NewTestSuite 创建新的测试套件
func NewTestSuite(t *testing.T, config *TestDBConfig) *TestSuite {
	db := SetupTestDatabase(t, config)
	log.Init(log.NewLogConfig())
	return &TestSuite{
		DB:    db,
		Store: store.NewStore(db),
		Ctx:   context.Background(),
	}
}

// Cleanup 清理测试套件
func (ts *TestSuite) Cleanup(t *testing.T) {
	CleanupTestDatabase(t, ts.DB)
}

// AssertUserExists 断言用户存在
func (ts *TestSuite) AssertUserExists(t *testing.T, email string) *model.UserM {
	user, err := ts.Store.User().Get(ts.Ctx, email)
	require.NoError(t, err, "user should exist")
	require.NotNil(t, user, "user should not be nil")
	return user
}

// AssertUserNotExists 断言用户不存在
func (ts *TestSuite) AssertUserNotExists(t *testing.T, email string) {
	_, err := ts.Store.User().Get(ts.Ctx, email)
	assert.Error(t, err, "user should not exist")
}

// AssertImageExists 断言图片存在
func (ts *TestSuite) AssertImageExists(t *testing.T, imageUUID string) *model.NewImageM {
	image, err := ts.Store.Image().Get(ts.Ctx, imageUUID)
	require.NoError(t, err, "image should exist")
	require.NotNil(t, image, "image should not be nil")
	return image
}

// AssertImageNotExists 断言图片不存在
func (ts *TestSuite) AssertImageNotExists(t *testing.T, imageUUID string) {
	_, err := ts.Store.Image().Get(ts.Ctx, imageUUID)
	assert.Error(t, err, "image should not exist")
}

// AssertImageHasTags 断言图片包含指定标签
func (ts *TestSuite) AssertImageHasTags(t *testing.T, imageUUID string, expectedTags []string) {
	image, err := ts.Store.Image().Get(ts.Ctx, imageUUID)
	require.NoError(t, err)

	actualTags := make([]string, len(image.Tags))
	for i, tag := range image.Tags {
		actualTags[i] = tag.Tag
	}

	for _, expectedTag := range expectedTags {
		assert.Contains(t, actualTags, expectedTag, "image should contain tag: %s", expectedTag)
	}
}

// AssertImageDoesNotHaveTag 断言图片不包含指定标签
func (ts *TestSuite) AssertImageDoesNotHaveTag(t *testing.T, imageUUID string, tag string) {
	image, err := ts.Store.Image().Get(ts.Ctx, imageUUID)
	require.NoError(t, err)

	for _, imageTag := range image.Tags {
		assert.NotEqual(t, tag, imageTag.Tag, "image should not contain tag: %s", tag)
	}
}

// WaitForCondition 等待条件满足
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-timeoutCh:
			t.Fatalf("timeout waiting for condition: %s", message)
		}
	}
}

// CountRecords 统计表中记录数
func (ts *TestSuite) CountRecords(t *testing.T, tableName string) int64 {
	var count int64
	err := ts.DB.Table(tableName).Count(&count).Error
	require.NoError(t, err, "failed to count records in table: %s", tableName)
	return count
}

// CountUsers 统计用户数
func (ts *TestSuite) CountUsers(t *testing.T) int64 {
	return ts.CountRecords(t, "users")
}

// CountImages 统计图片数
func (ts *TestSuite) CountImages(t *testing.T) int64 {
	return ts.CountRecords(t, "images")
}

// CountImageTags 统计图片标签数
func (ts *TestSuite) CountImageTags(t *testing.T) int64 {
	return ts.CountRecords(t, "image_tags")
}

// GetUserByEmail 通过邮箱获取用户
func (ts *TestSuite) GetUserByEmail(t *testing.T, email string) *model.UserM {
	user, err := ts.Store.User().Get(ts.Ctx, email)
	require.NoError(t, err)
	return user
}

// GetImageByUUID 通过UUID获取图片
func (ts *TestSuite) GetImageByUUID(t *testing.T, imageUUID string) *model.NewImageM {
	image, err := ts.Store.Image().Get(ts.Ctx, imageUUID)
	require.NoError(t, err)
	return image
}

// CreateUserAndAssert 创建用户并断言成功
func (ts *TestSuite) CreateUserAndAssert(t *testing.T, user *model.UserM) {
	err := ts.Store.User().Create(ts.Ctx, user)
	require.NoError(t, err, "failed to create user")

	// 验证用户已创建
	createdUser := ts.AssertUserExists(t, user.Email)
	assert.Equal(t, user.UserUUID, createdUser.UserUUID)
	assert.Equal(t, user.Email, createdUser.Email)
	assert.Equal(t, user.Nickname, createdUser.Nickname)
}

// CreateImageAndAssert 创建图片并断言成功
func (ts *TestSuite) CreateImageAndAssert(t *testing.T, image *model.NewImageM) {
	err := ts.Store.Image().Create(ts.Ctx, image)
	require.NoError(t, err, "failed to create image")
}

// DeleteUserAndAssert 删除用户并断言成功
func (ts *TestSuite) DeleteUserAndAssert(t *testing.T, userUUID string) {
	err := ts.Store.User().Delete(ts.Ctx, userUUID)
	require.NoError(t, err, "failed to delete user")
}

// DeleteImageAndAssert 删除图片并断言成功
func (ts *TestSuite) DeleteImageAndAssert(t *testing.T, imageUUID string) {
	err := ts.Store.Image().Delete(ts.Ctx, imageUUID)
	require.NoError(t, err, "failed to delete image")
}
