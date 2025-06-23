# Controller 集成测试

本目录包含了项目controller层的集成测试代码，用于测试HTTP API接口的完整功能。

## 测试结构

```
controller_test/
├── controller_integration_test.go  # 完整的集成测试套件
├── user_controller_test.go         # 用户controller专项测试
├── image_controller_test.go        # 图片controller专项测试
├── run_tests.sh                    # 测试运行脚本
├── coverage.out                    # 测试覆盖率数据（运行后生成）
├── coverage.html                   # 测试覆盖率报告（运行后生成）
└── README.md                       # 本文档
```

## 测试内容

### 用户Controller测试 (`user_controller_test.go`)

测试用户相关的API接口：

- **用户注册** (`POST /v1/auth/register`)
  - ✅ 成功注册
  - ✅ 验证错误（缺少必填字段、无效邮箱格式等）
  - ✅ 重复邮箱注册

- **用户登录** (`POST /v1/auth/login`)
  - ✅ 成功登录
  - ✅ 无效凭据
  - ✅ 用户不存在

- **获取用户信息** (`GET /v1/users/:user_uuid`)
  - ✅ 成功获取
  - ✅ 未授权访问
  - ✅ 用户不存在

- **更新用户信息** (`PATCH /v1/users/:user_uuid`)
  - ✅ 成功更新
  - ✅ 未授权访问
  - ✅ 无效数据

- **修改密码** (`POST /v1/auth/change-password`)
  - ✅ 成功修改
  - ✅ 旧密码错误
  - ✅ 未授权访问

### 图片Controller测试 (`image_controller_test.go`)

测试图片相关的API接口：

- **图片上传** (`POST /v1/images`)
  - ✅ 成功上传
  - ✅ 未授权上传
  - ✅ 无文件上传
  - ✅ 无效文件类型

- **获取公开图片列表** (`GET /v1/images`)
  - ✅ 成功获取
  - ✅ 分页参数测试

- **获取用户公开图片** (`GET /v1/images/user/:user_uuid/public`)
  - ✅ 成功获取
  - ✅ 无效用户UUID

- **获取图片详情** (`GET /v1/images/:image_uuid`)
  - ✅ 成功获取
  - ✅ 未授权访问
  - ✅ 图片不存在

- **获取用户图片列表** (`GET /v1/images/user/:user_uuid/images`)
  - ✅ 成功获取
  - ✅ 访问控制测试

- **更新图片标签** (`PATCH /v1/images/:image_uuid`)
  - ✅ 成功更新
  - ✅ 未授权更新
  - ✅ 无效标签

- **删除图片** (`DELETE /v1/images/:image_uuid`)
  - ✅ 成功删除
  - ✅ 未授权删除

- **获取图片文件** (`GET /v1/images/file/:imageUUIDFileName`)
  - ✅ 成功获取
  - ✅ 文件不存在

### 完整集成测试 (`controller_integration_test.go`)

包含跨controller的完整业务流程测试：
- 用户注册 → 登录 → 上传图片 → 管理图片的完整流程
- 多用户交互场景
- 权限控制验证

## 运行测试

### 方法一：使用测试脚本（推荐）

```bash
# 在项目根目录下运行
./test/controller_test/run_tests.sh
```

这个脚本会：
1. 检查测试环境
2. 运行所有controller测试
3. 生成测试覆盖率报告
4. 清理临时文件

### 方法二：手动运行

```bash
# 在项目根目录下

# 运行所有controller测试
go test -v ./test/controller_test

# 运行特定测试套件
go test -v ./test/controller_test -run TestUserController
go test -v ./test/controller_test -run TestImageController
go test -v ./test/controller_test -run TestControllerIntegration

# 生成覆盖率报告
go test -v ./test/controller_test -coverprofile=test/controller_test/coverage.out
go tool cover -html=test/controller_test/coverage.out -o test/controller_test/coverage.html
```

### 方法三：运行单个测试用例

```bash
# 运行特定的测试用例
go test -v ./test/controller_test -run TestUserController/TestCreate_Success
go test -v ./test/controller_test -run TestImageController/TestCreate_Success
```

## 测试环境要求

### 必需配置

1. **数据库配置**
   - 确保 `configs/demo520.yaml` 文件存在
   - 配置测试数据库连接信息
   - 测试会自动创建和清理测试数据

2. **依赖包**
   ```bash
   go mod tidy
   ```

3. **环境变量**
   ```bash
   export GO_ENV=test
   export GIN_MODE=test
   ```

### 可选配置

1. **测试图片目录**
   - 测试会自动创建临时目录
   - 可以通过环境变量 `TEST_IMAGE_DIR` 指定

2. **JWT密钥**
   - 测试使用固定的测试密钥
   - 生产环境请使用不同的密钥

## 测试数据管理

### 数据隔离
- 每个测试用例都有独立的数据环境
- 测试前自动清理数据库
- 测试后自动清理临时文件

### 测试用户
- 自动创建测试用户
- 使用faker生成随机数据
- 每次测试都是全新的用户数据

### 测试图片
- 自动创建测试图片记录
- 生成临时图片文件
- 测试完成后自动清理

## 常见问题

### 1. 数据库连接失败
```
Error: failed to connect to database
```
**解决方案：**
- 检查 `configs/demo520.yaml` 中的数据库配置
- 确保数据库服务正在运行
- 验证数据库用户权限

### 2. 测试超时
```
Error: test timeout
```
**解决方案：**
- 增加测试超时时间：`go test -timeout 30s`
- 检查数据库性能
- 优化测试用例

### 3. 权限错误
```
Error: permission denied
```
**解决方案：**
- 确保测试脚本有执行权限：`chmod +x run_tests.sh`
- 检查临时目录写入权限

### 4. 端口冲突
```
Error: address already in use
```
**解决方案：**
- 测试使用httptest，不需要真实端口
- 如果仍有问题，检查是否有其他服务占用端口

## 扩展测试

### 添加新的测试用例

1. **在现有测试文件中添加**
   ```go
   func (s *UserControllerTestSuite) TestNewFeature() {
       // 测试代码
   }
   ```

2. **创建新的测试文件**
   ```go
   // new_controller_test.go
   package controller_test
   
   import (
       "testing"
       "github.com/stretchr/testify/suite"
   )
   
   type NewControllerTestSuite struct {
       suite.Suite
       // 测试套件字段
   }
   
   func TestNewController(t *testing.T) {
       suite.Run(t, new(NewControllerTestSuite))
   }
   ```

### 性能测试

```go
func BenchmarkUserCreate(b *testing.B) {
    // 性能测试代码
}
```

### 并发测试

```go
func (s *UserControllerTestSuite) TestConcurrentAccess() {
    // 并发测试代码
}
```

## 最佳实践

1. **测试命名**
   - 使用描述性的测试名称
   - 格式：`Test<功能>_<场景>`
   - 例如：`TestCreate_Success`, `TestLogin_InvalidCredentials`

2. **测试结构**
   - 使用AAA模式：Arrange, Act, Assert
   - 每个测试只验证一个功能点
   - 使用子测试组织相关测试用例

3. **数据管理**
   - 使用faker生成测试数据
   - 避免硬编码测试数据
   - 确保测试数据的独立性

4. **错误处理**
   - 验证错误码和错误消息
   - 测试边界条件
   - 覆盖异常场景

5. **维护性**
   - 使用辅助函数减少重复代码
   - 保持测试代码的可读性
   - 定期更新测试用例

## 贡献指南

1. 添加新功能时，同时添加对应的测试用例
2. 确保测试覆盖率不低于80%
3. 运行完整测试套件确保没有回归
4. 更新相关文档

## 相关文档

- [项目README](../../README.md)
- [API文档](../../docs/api.md)
- [开发指南](../../docs/development.md)
- [部署指南](../../docs/deployment.md)