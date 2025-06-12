package biz_test

import (
	"context"
	"demo520/internal/520/biz"
	"demo520/internal/520/store"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/model"
	"demo520/pkg/api"
	"demo520/pkg/token"
	"fmt"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDatabase() (*gorm.DB, error) {
	// 3. 构造 DSN
	dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3316)/testdb?charset=utf8mb4&parseTime=True&loc=Local", "testpassword")

	// 4. 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.UserM{}); err != nil {
		return nil, err
	}

	return db, nil
}

func genNewUser(t *testing.T, db *gorm.DB, req *api.CreateUserRequest) (*api.CreateUserRequest, error) {
	iStore := store.NewStore(db)
	biz := biz.NewIBiz(iStore)
	userBiz := biz.Users()
	ctx := context.Background()
	var createReq api.CreateUserRequest
	if req == nil {
		createReq = api.CreateUserRequest{
			Nickname: faker.Name(),
			Email:    faker.Email(),
			Password: faker.Password(),
		}
	} else {
		createReq.Email = req.Email
		createReq.Nickname = req.Nickname
		createReq.Password = req.Password
	}
	err := userBiz.Create(ctx, &createReq)
	if err != nil {
		t.Fatal(err)
		return nil, err
	}
	return &createReq, nil
}

func TestUserBiz_Create_Success(t *testing.T) {
	db, err := setupDatabase()
	if err != nil {
		t.Error(err)
		return
	}
	ctx := context.Background()
	createReq, err := genNewUser(t, db, nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	iStore := store.NewStore(db)
	biz := biz.NewIBiz(iStore)
	userBiz := biz.Users()
	userResp, err := userBiz.Get(ctx, createReq.Email)
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.True(t, userResp.Email == createReq.Email)
	assert.True(t, userResp.Nickname == createReq.Nickname)
}

func TestUserBiz_Login_Success(t *testing.T) {
	db, err := setupDatabase()
	if err != nil {
		t.Fatalf("failed to setup database: %v", err)
	}
	userCreateReq := api.CreateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
		Password: faker.Password(),
	}
	_, err = genNewUser(t, db, &userCreateReq)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	iStore := store.NewStore(db)
	biz := biz.NewIBiz(iStore)
	userBiz := biz.Users()
	ctx := context.Background()
	loginReq := api.LoginRequest{
		Email:    userCreateReq.Email,
		Password: userCreateReq.Password,
		SeedTime: time.Now().Unix(),
	}
	userInfo, err := userBiz.Get(ctx, userCreateReq.Email)
	if err != nil {
		t.Fatalf("failed to get user info: %v", err)
	}
	userUUID := userInfo.UserUUID
	userResp, err := userBiz.Login(ctx, &loginReq)
	if err != nil {
		t.Fatalf("userBiz.Login failed: %v", err)
	}
	uuidFormToken, err := token.ParseToken(userResp.Token)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	assert.True(t, uuidFormToken.UserUUID == userUUID)
}

func TestUserChangePassword_Success(t *testing.T) {
	db, err := setupDatabase()
	if err != nil {
		t.Fatalf("failed to setup database: %v", err)
	}
	userCreateReq := api.CreateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
		Password: faker.Password(),
	}
	_, err = genNewUser(t, db, &userCreateReq)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	iStore := store.NewStore(db)
	biz := biz.NewIBiz(iStore)
	userBiz := biz.Users()
	ctx := context.Background()
	changePasswordReq := api.ChangePasswordRequest{
		OldPassword: userCreateReq.Password,
		NewPassword: faker.Password(),
	}
	if err := userBiz.ChangePassword(ctx, userCreateReq.Email, &changePasswordReq); err != nil {
		t.Fatalf("failed to change password: %v", err)
	}
	userInfo, err := userBiz.Get(ctx, userCreateReq.Email)
	if err != nil {
		t.Fatalf("failed to get user info: %v", err)
	}
	loginReq := api.LoginRequest{
		Email:    userCreateReq.Email,
		Password: userCreateReq.Password,
		SeedTime: time.Now().Unix(),
	}
	userUUID := userInfo.UserUUID
	_, err = userBiz.Login(ctx, &loginReq)
	if err != nil {
		if err != errno.ErrPasswordIncorrect {
			t.Fatalf("userBiz.Login failed: %v", err)
		}
	}
	loginReq.Password = changePasswordReq.NewPassword
	userResp, err := userBiz.Login(ctx, &loginReq)
	if err != nil {
		t.Fatalf("userBiz.Login failed: %v", err)
	}
	uuidFormToken, err := token.ParseToken(userResp.Token)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	assert.True(t, uuidFormToken.UserUUID == userUUID)
}

func TestUserBiz_Update_success(t *testing.T) {
	db, err := setupDatabase()
	if err != nil {
		t.Fatalf("failed to setup database: %v", err)
	}
	userCreateReq := api.CreateUserRequest{
		Email:    faker.Email(),
		Nickname: faker.Name(),
		Password: faker.Password(),
	}
	_, err = genNewUser(t, db, &userCreateReq)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	iStore := store.NewStore(db)
	biz := biz.NewIBiz(iStore)
	userBiz := biz.Users()
	ctx := context.Background()
	userInfo, err := userBiz.Get(ctx, userCreateReq.Email)
	if err != nil {
		t.Fatalf("failed to get user info: %v", err)
		return
	}
	updateReq := api.UpdateUserRequest{
		Nickname: faker.Name(),
		Email:    faker.Email(),
	}
	if err := userBiz.Update(ctx, userInfo.UserUUID, userCreateReq.Email, &updateReq); err != nil {
		t.Fatalf("failed to update user info: %v", err)
		return
	}
	_, err = userBiz.Get(ctx, userCreateReq.Email)
	if err != nil {
	}
	userResp, err := userBiz.Get(ctx, updateReq.Email)
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.True(t, userResp.Email == updateReq.Email)
	assert.True(t, userResp.Nickname == updateReq.Nickname)
}
