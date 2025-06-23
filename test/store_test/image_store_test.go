package store_test

import (
	"demo520/test/testhelper"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageStore_Create(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	imageStore := ts.Store.Image()

	t.Run("create valid image", func(t *testing.T) {
		// 创建测试用户
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)

		// 创建测试图片
		testImage := testhelper.CreateTestImage(t, ts.DB, testUser.User.UserUUID, nil)

		// 验证图片已创建
		imageUUIDStr := uuid.UUID(testImage.Image.ImageUUID).String()
		ts.AssertImageExists(t, imageUUIDStr)
		assert.Equal(t, int64(1), ts.CountImages(t))
	})

	t.Run("create image with nil pointer", func(t *testing.T) {
		err := imageStore.Create(ts.Ctx, nil)
		assert.Error(t, err)
	})
}

func TestImageStore_Get(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	imageStore := ts.Store.Image()

	t.Run("get existing image", func(t *testing.T) {
		// 创建测试用户和图片
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)
		testImage := testhelper.CreateTestImage(t, ts.DB, testUser.User.UserUUID, nil)

		// 获取图片
		imageUUIDStr := uuid.UUID(testImage.Image.ImageUUID).String()
		fetchedImage, err := imageStore.Get(ts.Ctx, imageUUIDStr)
		require.NoError(t, err)
		assert.Equal(t, testImage.Image.ImageUUID, fetchedImage.ImageUUID)
		assert.Equal(t, testImage.Image.UserUUID, fetchedImage.UserUUID)
	})

	t.Run("get non-existing image", func(t *testing.T) {
		_, err := imageStore.Get(ts.Ctx, uuid.New().String())
		assert.Error(t, err)
	})

	t.Run("get image with invalid UUID", func(t *testing.T) {
		_, err := imageStore.Get(ts.Ctx, "invalid-uuid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID format")
	})
}

func TestImageStore_Delete(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	imageStore := ts.Store.Image()

	t.Run("delete existing image", func(t *testing.T) {
		// 创建测试用户和图片
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)
		testImage := testhelper.CreateTestImage(t, ts.DB, testUser.User.UserUUID, nil)

		// 删除图片
		imageUUIDStr := uuid.UUID(testImage.Image.ImageUUID).String()
		ts.DeleteImageAndAssert(t, imageUUIDStr)

		// 验证图片已删除
		ts.AssertImageNotExists(t, imageUUIDStr)
	})

	t.Run("delete non-existing image", func(t *testing.T) {
		err := imageStore.Delete(ts.Ctx, uuid.New().String())
		// 删除不存在的图片不应该报错
		assert.NoError(t, err)
	})
}

func TestImageStore_AddTagsToImage(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	imageStore := ts.Store.Image()

	t.Run("add tags to image", func(t *testing.T) {
		// 创建测试用户和图片
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)
		testImage := testhelper.CreateTestImage(t, ts.DB, testUser.User.UserUUID, nil)

		// 添加标签
		tags := []string{"nature", "landscape", "photography"}
		imageUUIDStr := uuid.UUID(testImage.Image.ImageUUID).String()
		err := imageStore.AddTagsToImage(ts.Ctx, imageUUIDStr, tags)
		require.NoError(t, err)

		// 验证标签已添加
		ts.AssertImageHasTags(t, imageUUIDStr, tags)
		assert.Equal(t, int64(len(tags)), ts.CountImageTags(t))
	})

	t.Run("add duplicate tags", func(t *testing.T) {
		// 创建测试用户和图片
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)
		testImage := testhelper.CreateTestImage(t, ts.DB, testUser.User.UserUUID, nil)

		// 添加标签
		tags := []string{"nature", "nature", "landscape"}
		imageUUIDStr := uuid.UUID(testImage.Image.ImageUUID).String()
		err := imageStore.AddTagsToImage(ts.Ctx, imageUUIDStr, tags)
		require.NoError(t, err)

		// 验证重复标签只添加一次
		ts.AssertImageHasTags(t, imageUUIDStr, []string{"nature", "landscape"})
	})

	t.Run("add tags to non-existing image", func(t *testing.T) {
		tags := []string{"test"}
		err := imageStore.AddTagsToImage(ts.Ctx, uuid.New().String(), tags)
		assert.Error(t, err)
	})
}

func TestImageStore_DeleteTagFromImage(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	imageStore := ts.Store.Image()

	t.Run("delete tag from image", func(t *testing.T) {
		// 创建带标签的测试图片
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)
		tags := []string{"nature", "landscape", "photography"}
		testImage := testhelper.CreateTestImageWithTags(t, ts.DB, testUser.User.UserUUID, tags, false)

		// 删除一个标签
		imageUUIDStr := uuid.UUID(testImage.Image.ImageUUID).String()
		err := imageStore.DeleteTagFromImage(ts.Ctx, imageUUIDStr, []string{"landscape"})
		require.NoError(t, err)

		// 验证标签已删除
		ts.AssertImageDoesNotHaveTag(t, imageUUIDStr, "landscape")
		ts.AssertImageHasTags(t, imageUUIDStr, []string{"nature", "photography"})
	})

	t.Run("delete non-existing tag", func(t *testing.T) {
		// 创建带标签的测试图片
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)
		tags := []string{"nature"}
		testImage := testhelper.CreateTestImageWithTags(t, ts.DB, testUser.User.UserUUID, tags, false)

		// 尝试删除不存在的标签
		imageUUIDStr := uuid.UUID(testImage.Image.ImageUUID).String()
		err := imageStore.DeleteTagFromImage(ts.Ctx, imageUUIDStr, []string{"nonexistent"})
		// 删除不存在的标签不应该报错
		assert.NoError(t, err)
	})
}

func TestImageStore_GetUserImages(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	imageStore := ts.Store.Image()

	t.Run("get user images with pagination", func(t *testing.T) {
		// 创建测试用户
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)

		// 创建多个测试图片
		testImages := testhelper.CreateTestImages(t, ts.DB, testUser.User.UserUUID, 5)

		// 获取用户图片
		count, images, err := imageStore.GetUserImages(ts.Ctx, testUser.User.UserUUID, 0, 3)
		require.NoError(t, err)
		assert.Equal(t, int64(len(testImages)), count)
		assert.Len(t, images, 3)

		// 获取第二页
		count, images, err = imageStore.GetUserImages(ts.Ctx, testUser.User.UserUUID, 3, 3)
		require.NoError(t, err)
		assert.Equal(t, int64(len(testImages)), count)
		assert.Len(t, images, 2)
	})

	t.Run("get images for user with no images", func(t *testing.T) {
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)

		count, images, err := imageStore.GetUserImages(ts.Ctx, testUser.User.UserUUID, 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
		assert.Len(t, images, 0)
	})

	t.Run("get images for non-existing user", func(t *testing.T) {
		count, images, err := imageStore.GetUserImages(ts.Ctx, uuid.New().String(), 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
		assert.Len(t, images, 0)
	})
}

func TestImageStore_GetRandomPublicImages(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	imageStore := ts.Store.Image()

	t.Run("get random public images", func(t *testing.T) {
		// 创建测试用户
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)
		defer teardownStoreTest(t, ts)

		// 创建公开和私有图片
		publicTags := []string{"public"}
		privateTags := []string{"private"}

		// 创建3个公开图片
		for i := 0; i < 3; i++ {
			testhelper.CreateTestImageWithTags(t, ts.DB, testUser.User.UserUUID, publicTags, true)
		}

		// 创建2个私有图片
		for i := 0; i < 2; i++ {
			testhelper.CreateTestImageWithTags(t, ts.DB, testUser.User.UserUUID, privateTags, false)
		}

		// 获取随机公开图片
		count, images, err := imageStore.GetRandomPublicImages(ts.Ctx, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count) // 只有3个公开图片
		assert.Len(t, images, 3)

		// 验证所有返回的图片都是公开的
		for _, image := range images {
			assert.True(t, image.IsPublic)
		}
	})

	t.Run("get random public images with limit", func(t *testing.T) {
		// 创建测试用户
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)
		defer teardownStoreTest(t, ts)

		// 创建5个公开图片
		publicTags := []string{"public"}
		for i := 0; i < 5; i++ {
			testhelper.CreateTestImageWithTags(t, ts.DB, testUser.User.UserUUID, publicTags, true)
		}

		// 限制返回2个图片
		count, images, err := imageStore.GetRandomPublicImages(ts.Ctx, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(5), count) // 总共5个公开图片
		assert.Len(t, images, 2)         // 但只返回2个
	})

	t.Run("get random public images when no public images exist", func(t *testing.T) {
		// 创建测试用户和私有图片
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)
		testhelper.CreateTestImageWithTags(t, ts.DB, testUser.User.UserUUID, []string{"private"}, false)
		defer teardownStoreTest(t, ts)

		count, images, err := imageStore.GetRandomPublicImages(ts.Ctx, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
		assert.Len(t, images, 0)
	})
}
