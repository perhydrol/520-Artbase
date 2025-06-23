package biz_test

import (
	"context"
	"demo520/internal/520/biz/user"
	"demo520/internal/pkg/errno"
	"demo520/pkg/api"
	"demo520/pkg/token"
	"demo520/test/testhelper"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupBizTest 设置biz层测试环境
func setupBizTest(t *testing.T) (*testhelper.TestSuite, user.UserBiz) {
	ts := testhelper.NewTestSuite(t, nil)
	userBiz := user.NewUserBiz(ts.Store)
	return ts, userBiz
}

// teardownBizTest 清理biz层测试环境
func teardownBizTest(ts *testhelper.TestSuite, t *testing.T) {
	ts.Cleanup(t)
}

// createTestUserViaBiz 通过biz层创建测试用户
func createTestUserViaBiz(t *testing.T, userBiz user.UserBiz, req *api.CreateUserRequest) *api.UserInfo {
	ctx := context.Background()
	var createReq api.CreateUserRequest
	if req == nil {
		createReq = api.CreateUserRequest{
			Nickname: faker.Name(),
			Email:    faker.Email(),
			Password: faker.Password(),
		}
	} else {
		createReq = *req
	}
	err := userBiz.Create(ctx, &createReq)
	require.NoError(t, err)
	getResp, err := userBiz.Get(ctx, createReq.Email)
	require.NoError(t, err)
	userInfo := api.UserInfo(*getResp)
	return &userInfo
}

func TestUserBiz_Create_Success(t *testing.T) {
	ts, userBiz := setupBizTest(t)
	defer teardownBizTest(ts, t)

	userCreateReq := &api.CreateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
		Password: faker.Password(),
	}
	createTestUserViaBiz(t, userBiz, userCreateReq)

	ctx := context.Background()
	loginReq := api.LoginRequest{
		Email:    userCreateReq.Email,
		Password: userCreateReq.Password,
		SeedTime: time.Now().Unix(),
	}

	userInfo, err := userBiz.Get(ctx, userCreateReq.Email)
	require.NoError(t, err)
	userUUID := userInfo.UserUUID

	userResp, err := userBiz.Login(ctx, &loginReq)
	require.NoError(t, err)

	uuidFormToken, err := token.ParseToken(userResp.Token)
	require.NoError(t, err)

	assert.Equal(t, userUUID, uuidFormToken.UserUUID)
}

func TestUserBiz_ChangePassword_Success(t *testing.T) {
	ts, userBiz := setupBizTest(t)
	defer teardownBizTest(ts, t)

	userCreateReq := &api.CreateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
		Password: faker.Password(),
	}
	createTestUserViaBiz(t, userBiz, userCreateReq)

	ctx := context.Background()
	changePasswordReq := api.ChangePasswordRequest{
		OldPassword: userCreateReq.Password,
		NewPassword: faker.Password(),
	}

	err := userBiz.ChangePassword(ctx, userCreateReq.Email, &changePasswordReq)
	require.NoError(t, err)

	userInfo, err := userBiz.Get(ctx, userCreateReq.Email)
	require.NoError(t, err)
	userUUID := userInfo.UserUUID

	// 验证旧密码不能登录
	oldLoginReq := api.LoginRequest{
		Email:    userCreateReq.Email,
		Password: userCreateReq.Password,
		SeedTime: time.Now().Unix(),
	}
	_, err = userBiz.Login(ctx, &oldLoginReq)
	assert.Error(t, err)
	assert.Equal(t, errno.ErrPasswordIncorrect, err)

	// 验证新密码可以登录
	newLoginReq := api.LoginRequest{
		Email:    userCreateReq.Email,
		Password: changePasswordReq.NewPassword,
		SeedTime: time.Now().Unix(),
	}
	userResp, err := userBiz.Login(ctx, &newLoginReq)
	require.NoError(t, err)

	uuidFormToken, err := token.ParseToken(userResp.Token)
	require.NoError(t, err)

	assert.Equal(t, userUUID, uuidFormToken.UserUUID)
}

func TestUserBiz_Update_Success(t *testing.T) {
	ts, userBiz := setupBizTest(t)
	defer teardownBizTest(ts, t)

	userCreateReq := &api.CreateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
		Password: faker.Password(),
	}
	createTestUserViaBiz(t, userBiz, userCreateReq)

	ctx := context.Background()
	userInfo, err := userBiz.Get(ctx, userCreateReq.Email)
	require.NoError(t, err)

	updateReq := api.UpdateUserRequest{
		Nickname: faker.Name(),
		Email:    faker.Email(),
	}

	err = userBiz.Update(ctx, userInfo.UserUUID, userCreateReq.Email, &updateReq)
	require.NoError(t, err)

	// 验证旧邮箱不存在
	_, err = userBiz.Get(ctx, userCreateReq.Email)
	assert.Error(t, err)

	// 验证新邮箱存在且信息正确
	userResp, err := userBiz.Get(ctx, updateReq.Email)
	require.NoError(t, err)

	assert.Equal(t, updateReq.Email, userResp.Email)
	assert.Equal(t, updateReq.Nickname, userResp.Nickname)
	assert.Equal(t, userInfo.UserUUID, userResp.UserUUID)
}
