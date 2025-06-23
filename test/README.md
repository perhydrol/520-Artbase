# 测试框架使用指南

本项目提供了完整的集成测试框架，包含数据库连接、测试数据生成和断言辅助函数。

## 目录结构

```
test/
├── testhelper/           # 测试辅助函数包
│   ├── database.go       # 数据库连接和迁移
│   ├── fixtures.go       # 测试数据生成
│   └── helpers.go        # 通用测试辅助函数
├── store_test/           # Store层测试
│   ├── store_helper_test.go   # Store测试辅助函数
│   ├── user_test.go      # 用户Store基础测试
│   ├── user_store_test.go # 用户Store详细测试
│   ├── image_test.go     # 图片集成测试
│   └── image_store_test.go # 图片Store详细测试
├── controller_test/      # Controller层测试
├── biz_test/            # 业务逻辑测试
└── README.md            # 本文件
```

## 快速开始

### 1. 数据库配置

测试使用独立的测试数据库，默认配置：
- Host: 127.0.0.1:3316
- Username: root
- Password: testpassword
- Database: testdb

可以通过 `testhelper.TestDBConfig` 自定义配置。

### 2. 基本测试模式

```go
func TestExample(t *testing.T) {
    // 设置测试环境
    ts := testhelper.NewTestSuite(t, nil) // 使用默认配置
    defer ts.Cleanup(t)
    
    // 创建测试数据
    testUser := testhelper.CreateTestUser(t, ts.DB, nil)
    
    // 执行测试逻辑
    // ...
    
    // 使用断言函数验证结果
    ts.AssertUserExists(t, testUser.Email)
}
```

### 3. Store层测试模式

```go
func TestStoreExample(t *testing.T) {
    // 使用Store测试辅助函数
    ts := setupStoreTest(t)
    defer teardownStoreTest(t, ts)
    
    // 获取Store实例
    userStore := ts.Store.User()
    
    // 执行测试
    // ...
}
```

## 核心组件

### TestSuite

`TestSuite` 是测试的核心结构，包含：
- `DB`: GORM数据库实例
- `Store`: Store层接口实例
- `Ctx`: 上下文对象

### 数据库管理

- `SetupTestDatabase()`: 设置测试数据库连接
- `AutoMigrate()`: 执行数据库迁移
- `CleanupTestDatabase()`: 清理测试数据
- `TruncateTestDatabase()`: 截断所有表

### 测试数据生成

#### 用户数据
```go
// 创建单个测试用户
testUser := testhelper.CreateTestUser(t, db, nil)

// 创建自定义用户
customUser := &model.UserM{
    Email: "test@example.com",
    Password: "password123",
}
testUser := testhelper.CreateTestUser(t, db, customUser)

// 批量创建用户
testUsers := testhelper.CreateTestUsers(t, db, 5)
```

#### 图片数据
```go
// 创建基础图片
testImage := testhelper.CreateTestImage(t, db, userUUID, nil)

// 创建带标签的图片
tags := []string{"nature", "landscape"}
testImage := testhelper.CreateTestImageWithTags(t, db, userUUID, tags, true)

// 批量创建图片
testImages := testhelper.CreateTestImages(t, db, userUUID, 3)
```

### 断言函数

#### 用户断言
```go
// 断言用户存在
user := ts.AssertUserExists(t, email)

// 断言用户不存在
ts.AssertUserNotExists(t, email)
```

#### 图片断言
```go
// 断言图片存在
image := ts.AssertImageExists(t, imageUUID)

// 断言图片不存在
ts.AssertImageNotExists(t, imageUUID)

// 断言图片包含标签
ts.AssertImageHasTags(t, imageUUID, []string{"tag1", "tag2"})

// 断言图片不包含标签
ts.AssertImageDoesNotHaveTag(t, imageUUID, "tag")
```

#### 统计断言
```go
// 统计记录数
count := ts.CountUsers(t)
count := ts.CountImages(t)
count := ts.CountImageTags(t)
count := ts.CountRecords(t, "table_name")
```

### 便捷操作函数

```go
// 创建并断言成功
ts.CreateUserAndAssert(t, user)
ts.CreateImageAndAssert(t, image)

// 删除并断言成功
ts.DeleteUserAndAssert(t, userUUID)
ts.DeleteImageAndAssert(t, imageUUID)

// 获取数据
user := ts.GetUserByEmail(t, email)
image := ts.GetImageByUUID(t, imageUUID)
```

## 测试最佳实践

### 1. 测试隔离
每个测试函数都应该：
- 使用独立的测试环境
- 在测试结束后清理数据
- 不依赖其他测试的状态

### 2. 数据准备
```go
func TestExample(t *testing.T) {
    ts := setupStoreTest(t)
    defer teardownStoreTest(t, ts)
    
    // 准备测试数据
    testUser := testhelper.CreateTestUser(t, ts.DB, nil)
    
    // 执行测试逻辑
    // ...
}
```

### 3. 子测试组织
```go
func TestUserStore_Create(t *testing.T) {
    ts := setupStoreTest(t)
    defer teardownStoreTest(t, ts)
    
    t.Run("create valid user", func(t *testing.T) {
        // 测试正常情况
    })
    
    t.Run("create user with invalid data", func(t *testing.T) {
        // 测试异常情况
    })
}
```

### 4. 错误处理测试
```go
t.Run("handle error case", func(t *testing.T) {
    err := someOperation()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expected error message")
})
```

## 运行测试

```bash
# 运行所有测试
go test ./test/...

# 运行特定包的测试
go test ./test/store_test/

# 运行特定测试函数
go test ./test/store_test/ -run TestUserStore_Create

# 显示详细输出
go test -v ./test/store_test/

# 运行测试并显示覆盖率
go test -cover ./test/store_test/
```

## 注意事项

1. **数据库连接**: 确保测试数据库服务正在运行
2. **数据清理**: 每个测试都会自动清理数据，无需手动处理
3. **并发安全**: 测试框架支持并发测试，但建议使用不同的数据库实例
4. **性能考虑**: 大量数据的测试应该使用批量操作函数
5. **错误处理**: 所有测试辅助函数都会在失败时自动调用 `t.Fatal()`

## 扩展测试框架

如需添加新的测试辅助函数，请遵循以下原则：

1. 将通用函数放在 `testhelper` 包中
2. 将特定层的函数放在对应的测试包中
3. 使用一致的命名约定
4. 提供充分的错误处理
5. 添加适当的文档注释