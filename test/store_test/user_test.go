package store_test

import (
	"demo520/internal/pkg/model"
	"demo520/pkg/auth"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {
	// 设置测试环境
	ts := setupStoreTest(t)
	defer teardownStoreTest(t, ts)

	userStore := ts.Store.User()

	// 创建用户
	password := faker.Password()
	user := &model.UserM{
		UserUUID: uuid.New().String(),
		Password: password,
		Nickname: faker.Name(),
		Email:    faker.Email(),
	}

	// 创建用户并断言成功
	ts.CreateUserAndAssert(t, user)

	// 获取用户
	fetchedUser := ts.GetUserByEmail(t, user.Email)
	assert.Equal(t, user.UserUUID, fetchedUser.UserUUID)
	assert.True(t, auth.VerifyPassword(password, fetchedUser.Password))

	// 更新用户
	fetchedUser.Nickname = "UpdatedName"
	err := userStore.Update(ts.Ctx, fetchedUser)
	require.NoError(t, err)

	// 验证更新
	updatedUser := ts.GetUserByEmail(t, user.Email)
	assert.Equal(t, "UpdatedName", updatedUser.Nickname)

	// 列出用户
	users, err := userStore.List(ts.Ctx, 0, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(*users), 1)

	// 删除用户
	ts.DeleteUserAndAssert(t, user.UserUUID)

	// 验证用户已删除
	ts.AssertUserNotExists(t, user.Email)
}
