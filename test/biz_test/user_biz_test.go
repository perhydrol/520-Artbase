package biz_test

import (
	"context"
	"demo520/internal/pkg/errno"
	"demo520/pkg/api"
	"demo520/pkg/token"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserBiz_Create_ErrorCases 测试用户创建的错误情况
func TestUserBiz_Create_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		req     *api.CreateUserRequest
		wantErr error
	}{
		{
			name: "空邮箱",
			req: &api.CreateUserRequest{
				Email:    "",
				Nickname: faker.Name(),
				Password: faker.Password(),
			},
			wantErr: nil, // 具体错误类型需要根据实际实现确定
		},
		{
			name: "无效邮箱格式",
			req: &api.CreateUserRequest{
				Email:    "invalid-email",
				Nickname: faker.Name(),
				Password: faker.Password(),
			},
			wantErr: nil, // 具体错误类型需要根据实际实现确定
		},
		{
			name: "空密码",
			req: &api.CreateUserRequest{
				Email:    faker.Email(),
				Nickname: faker.Name(),
				Password: "",
			},
			wantErr: nil, // 具体错误类型需要根据实际实现确定
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, userBiz := setupBizTest(t)
			defer teardownBizTest(ts, t)

			ctx := context.Background()
			err := userBiz.Create(ctx, tt.req)
			assert.Error(t, err)
			// 如果有具体的错误类型，可以进一步验证
			// assert.Equal(t, tt.wantErr, err)
		})
	}
}

// TestUserBiz_Create_DuplicateEmail 测试重复邮箱创建用户
func TestUserBiz_Create_DuplicateEmail(t *testing.T) {
	ts, userBiz := setupBizTest(t)
	defer teardownBizTest(ts, t)

	email := faker.Email()
	userReq := &api.CreateUserRequest{
		Email:    email,
		Nickname: faker.Name(),
		Password: faker.Password(),
	}

	// 创建第一个用户
	createTestUserViaBiz(t, userBiz, userReq)

	// 尝试创建相同邮箱的用户
	duplicateReq := &api.CreateUserRequest{
		Email:    email,
		Nickname: faker.Name(),
		Password: faker.Password(),
	}

	ctx := context.Background()
	err := userBiz.Create(ctx, duplicateReq)
	assert.Error(t, err)
	assert.Equal(t, errno.ErrUserAlreadyExist, err)
}

// TestUserBiz_Get_NotFound 测试获取不存在的用户
func TestUserBiz_Get_NotFound(t *testing.T) {
	ts, userBiz := setupBizTest(t)
	defer teardownBizTest(ts, t)

	ctx := context.Background()
	_, err := userBiz.Get(ctx, "nonexistent@example.com")
	assert.Error(t, err)
	assert.Equal(t, errno.ErrUserNotFound, err)
}

// TestUserBiz_Login_ErrorCases 测试登录的错误情况
func TestUserBiz_Update_ErrorCases(t *testing.T) {
	ts, userBiz := setupBizTest(t)
	defer teardownBizTest(ts, t)

	userCreateReq := &api.CreateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
		Password: faker.Password(),
	}
	createTestUserViaBiz(t, userBiz, userCreateReq)

	tests := []struct {
		name    string
		req     *api.LoginRequest
		wantErr error
	}{
		{
			name: "错误密码",
			req: &api.LoginRequest{
				Email:    userCreateReq.Email,
				Password: "wrong-password",
				SeedTime: time.Now().Unix(),
			},
			wantErr: errno.ErrPasswordIncorrect,
		},
		{
			name: "不存在的用户",
			req: &api.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: faker.Password(),
				SeedTime: time.Now().Unix(),
			},
			wantErr: errno.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := userBiz.Login(ctx, tt.req)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

// TestUserBiz_ChangePassword_ErrorCases 测试修改密码的错误情况
func TestUserBiz_ChangePassword_ErrorCases(t *testing.T) {
	ts, userBiz := setupBizTest(t)
	defer teardownBizTest(ts, t)

	userCreateReq := &api.CreateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
		Password: faker.Password(),
	}
	createTestUserViaBiz(t, userBiz, userCreateReq)

	tests := []struct {
		name    string
		email   string
		req     *api.ChangePasswordRequest
		wantErr error
	}{
		{
			name:  "错误的旧密码",
			email: userCreateReq.Email,
			req: &api.ChangePasswordRequest{
				OldPassword: "wrong-old-password",
				NewPassword: faker.Password(),
			},
			wantErr: errno.ErrPasswordIncorrect,
		},
		{
			name:  "不存在的用户",
			email: "nonexistent@example.com",
			req: &api.ChangePasswordRequest{
				OldPassword: faker.Password(),
				NewPassword: faker.Password(),
			},
			wantErr: errno.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := userBiz.ChangePassword(ctx, tt.email, tt.req)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

// TestUserBiz_Update_ErrorCases 测试更新用户的错误情况
func TestUserBiz_Update_Error(t *testing.T) {
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

	tests := []struct {
		name     string
		userUUID string
		email    string
		req      *api.UpdateUserRequest
		wantErr  bool
	}{
		{
			name:     "不存在的用户UUID",
			userUUID: "non-existent-uuid",
			email:    userCreateReq.Email,
			req: &api.UpdateUserRequest{
				Nickname: faker.Name(),
				Email:    faker.Email(),
			},
			wantErr: true,
		},
		{
			name:     "错误的邮箱",
			userUUID: userInfo.UserUUID,
			email:    "wrong@example.com",
			req: &api.UpdateUserRequest{
				Nickname: faker.Name(),
				Email:    faker.Email(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := userBiz.Update(ctx, tt.userUUID, tt.email, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUserBiz_Delete_Success 测试删除用户成功
func TestUserBiz_Delete_Success(t *testing.T) {
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

	// 删除用户
	err = userBiz.Delete(ctx, userInfo.UserUUID)
	require.NoError(t, err)

	// 验证用户已被删除
	_, err = userBiz.Get(ctx, userCreateReq.Email)
	assert.Error(t, err)
	assert.Equal(t, errno.ErrUserNotFound, err)
}

// TestUserBiz_Delete_NotFound 测试删除不存在的用户
func TestUserBiz_Delete_NotFound(t *testing.T) {
	ts, userBiz := setupBizTest(t)
	defer teardownBizTest(ts, t)

	ctx := context.Background()
	err := userBiz.Delete(ctx, "non-existent-uuid")
	assert.Error(t, err)
}

// TestUserBiz_TokenValidation 测试Token验证
func TestUserBiz_Get_InvalidEmail(t *testing.T) {
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

	userResp, err := userBiz.Login(ctx, &loginReq)
	require.NoError(t, err)
	require.NotEmpty(t, userResp.Token)

	// 验证Token可以正确解析
	uuidFromToken, err := token.ParseToken(userResp.Token)
	require.NoError(t, err)

	// 获取用户信息验证UUID匹配
	userInfo, err := userBiz.Get(ctx, userCreateReq.Email)
	require.NoError(t, err)

	assert.Equal(t, userInfo.UserUUID, uuidFromToken.UserUUID)
}
