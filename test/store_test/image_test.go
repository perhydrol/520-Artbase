package store_test

import (
	"demo520/test/testhelper"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageIntegration(t *testing.T) {
	// 设置测试环境
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	// 创建测试用户
	testUsers := testhelper.CreateTestUsers(t, ts.DB, 3)
	require.Len(t, testUsers, 3)

	// 验证用户已创建
	for _, testUser := range testUsers {
		ts.AssertUserExists(t, testUser.Email)
	}

	// 为每个用户创建图片
	for i, testUser := range testUsers {
		// 创建带标签的图片
		tags := testhelper.GenerateTestTags(2)
		isPublic := i%2 == 0 // 交替创建公开和私有图片
		testImage := testhelper.CreateTestImageWithTags(t, ts.DB, testUser.User.UserUUID, tags, isPublic)

		// 验证图片已创建
		imageUUIDStr := testImage.Image.ImageUUID.String()
		ts.AssertImageExists(t, imageUUIDStr)
		ts.AssertImageHasTags(t, imageUUIDStr, tags)
	}

	// 验证总数
	assert.Equal(t, int64(3), ts.CountUsers(t))
	assert.Equal(t, int64(3), ts.CountImages(t))
	assert.GreaterOrEqual(t, ts.CountImageTags(t), int64(6)) // 至少6个标签
}
