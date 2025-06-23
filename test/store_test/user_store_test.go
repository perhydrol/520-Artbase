package store_test

import (
	"demo520/internal/pkg/model"
	"demo520/pkg/auth"
	"demo520/test/testhelper"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStore_Create(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	userStore := ts.Store.User()

	t.Run("create valid user", func(t *testing.T) {
		user := &model.UserM{
			UserUUID: uuid.New().String(),
			Password: faker.Password(),
			Nickname: faker.Name(),
			Email:    faker.Email(),
		}

		ts.CreateUserAndAssert(t, user)
		assert.Equal(t, int64(1), ts.CountUsers(t))
	})

	t.Run("create user with nil pointer", func(t *testing.T) {
		err := userStore.Create(ts.Ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user cannot be nil")
	})

	t.Run("create user with duplicate email", func(t *testing.T) {
		email := faker.Email()
		user1 := &model.UserM{
			UserUUID: uuid.New().String(),
			Password: faker.Password(),
			Nickname: faker.Name(),
			Email:    email,
		}
		user2 := &model.UserM{
			UserUUID: uuid.New().String(),
			Password: faker.Password(),
			Nickname: faker.Name(),
			Email:    email,
		}

		ts.CreateUserAndAssert(t, user1)
		err := userStore.Create(ts.Ctx, user2)
		assert.Error(t, err)
	})
}

func TestUserStore_Get(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	userStore := ts.Store.User()

	t.Run("get existing user", func(t *testing.T) {
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)

		fetchedUser, err := userStore.Get(ts.Ctx, testUser.Email)
		require.NoError(t, err)
		assert.Equal(t, testUser.User.UserUUID, fetchedUser.UserUUID)
		assert.Equal(t, testUser.User.Email, fetchedUser.Email)
	})

	t.Run("get non-existing user", func(t *testing.T) {
		_, err := userStore.Get(ts.Ctx, "nonexistent@example.com")
		assert.Error(t, err)
	})
}

func TestUserStore_Update(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	userStore := ts.Store.User()

	t.Run("update existing user", func(t *testing.T) {
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)

		// 更新用户信息
		testUser.User.Nickname = "Updated Nickname"
		err := userStore.Update(ts.Ctx, testUser.User)
		require.NoError(t, err)

		// 验证更新
		updatedUser := ts.GetUserByEmail(t, testUser.Email)
		assert.Equal(t, "Updated Nickname", updatedUser.Nickname)
	})

	t.Run("update user with nil pointer", func(t *testing.T) {
		err := userStore.Update(ts.Ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user cannot be nil")
	})

	t.Run("update user with empty UUID", func(t *testing.T) {
		user := &model.UserM{
			UserUUID: "",
			Email:    faker.Email(),
			Nickname: faker.Name(),
		}
		err := userStore.Update(ts.Ctx, user)
		assert.Error(t, err)
	})
}

func TestUserStore_Delete(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	userStore := ts.Store.User()

	t.Run("delete existing user", func(t *testing.T) {
		testUser := testhelper.CreateTestUser(t, ts.DB, nil)

		// 删除用户
		ts.DeleteUserAndAssert(t, testUser.User.UserUUID)

		// 验证用户已删除
		ts.AssertUserNotExists(t, testUser.Email)
	})

	t.Run("delete non-existing user", func(t *testing.T) {
		err := userStore.Delete(ts.Ctx, uuid.New().String())
		// 删除不存在的用户不应该报错
		assert.NoError(t, err)
	})
}

func TestUserStore_List(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	userStore := ts.Store.User()

	t.Run("list users with pagination", func(t *testing.T) {
		// 创建多个测试用户
		testUsers := testhelper.CreateTestUsers(t, ts.DB, 5)
		defer func() {
			for _, user := range testUsers {
				ts.DeleteUserAndAssert(t, user.User.UserUUID)
			}
		}()

		// 测试分页
		users, err := userStore.List(ts.Ctx, 0, 3)
		require.NoError(t, err)
		assert.Len(t, *users, 3)

		// 测试第二页
		users, err = userStore.List(ts.Ctx, 3, 3)
		require.NoError(t, err)
		assert.Len(t, *users, 2)

		// 验证总数
		assert.Equal(t, int64(len(testUsers)), ts.CountUsers(t))
	})

	t.Run("list empty users", func(t *testing.T) {
		users, err := userStore.List(ts.Ctx, 0, 10)
		require.NoError(t, err)
		assert.Len(t, *users, 0)
	})
}

func TestUserStore_ChangePassword(t *testing.T) {
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	userStore := ts.Store.User()

	t.Run("change password successfully", func(t *testing.T) {
		oldPassword := faker.Password()
		newPassword := faker.Password()

		testUser := testhelper.CreateTestUser(t, ts.DB, &model.UserM{
			Password: oldPassword,
		})

		// 更改密码
		err := userStore.ChangePassword(ts.Ctx, testUser.Email, oldPassword, newPassword)
		require.NoError(t, err)

		// 验证新密码
		updatedUser := ts.GetUserByEmail(t, testUser.Email)
		assert.True(t, auth.VerifyPassword(newPassword, updatedUser.Password))
		assert.False(t, auth.VerifyPassword(oldPassword, updatedUser.Password))
	})

	t.Run("change password with wrong old password", func(t *testing.T) {
		oldPassword := faker.Password()
		wrongPassword := faker.Password()
		newPassword := faker.Password()

		testUser := testhelper.CreateTestUser(t, ts.DB, &model.UserM{
			Password: oldPassword,
		})

		// 尝试用错误的旧密码更改密码
		err := userStore.ChangePassword(ts.Ctx, testUser.Email, wrongPassword, newPassword)
		assert.Error(t, err)
	})

	t.Run("change password for non-existing user", func(t *testing.T) {
		err := userStore.ChangePassword(ts.Ctx, "nonexistent@example.com", "oldpass", "newpass")
		assert.Error(t, err)
	})
}
