# 配置管理文档

## 概述

本项目使用集中化的配置管理方案，基于 Viper 库实现，提供了类型安全的配置访问和默认值管理。

## 配置结构

配置文件采用 YAML 格式，支持以下配置项：

### 服务器配置 (server)
- `addr`: 服务监听地址，默认 `:8080`

### 数据库配置 (db)
- `host`: 数据库主机地址，默认 `127.0.0.1:3306`
- `username`: 数据库用户名，默认 `root`
- `password`: 数据库密码，默认为空
- `database`: 数据库名称，默认 `demo520`
- `max-idle-connections`: 最大空闲连接数，默认 `10`
- `max-open-connections`: 最大打开连接数，默认 `100`
- `max-connection-life-time`: 连接最大生存时间，默认 `10s`
- `log-level`: 数据库日志级别，默认 `1`

### 日志配置 (log)
- `disable-caller`: 是否禁用调用者信息，默认 `false`
- `disable-stacktrace`: 是否禁用堆栈跟踪，默认 `false`
- `level`: 日志级别，默认 `info`
- `encoding`: 日志编码格式，默认 `console`
- `output-paths`: 日志输出路径，默认 `["stdout"]`

### 图片处理配置 (image)
- `ImageMaxSize`: 图片最大大小（字节），默认 `20971520` (20MB)
- `image_dir`: 图片存储目录，默认 `temp_image`
- `Convert`: 图片转换配置
  - `WebPQuality`: WebP 质量 (0-100)，默认 `80`
  - `WebReductionEffort`: WebP 压缩努力程度 (0-6)，默认 `4`
  - `AvifQuality`: AVIF 质量 (0-100)，默认 `60`
  - `AvifEffort`: AVIF 压缩努力程度 (0-9)，默认 `4`
  - `ImageLossless`: 是否使用无损压缩，默认 `false`

### JWT 配置 (jwt)
- `secret`: JWT 签名密钥，默认 `demo520-secret-key`

### 应用配置 (app)
- `runmode`: 运行模式，默认 `debug`

## 使用方法

### 1. 初始化配置

在应用启动时调用配置初始化：

```go
import "demo520/internal/pkg/config"

// 使用默认配置文件路径
if err := config.Init(""); err != nil {
    log.Fatal("配置初始化失败:", err)
}

// 或指定配置文件路径
if err := config.Init("/path/to/config.yaml"); err != nil {
    log.Fatal("配置初始化失败:", err)
}
```

### 2. 获取配置

使用类型安全的配置获取方法：

```go
// 获取完整配置
cfg := config.GetConfig()

// 获取特定模块配置
serverConfig := config.GetServer()
dbConfig := config.GetDB()
logConfig := config.GetLog()
imageConfig := config.GetImage()
jwtConfig := config.GetJWT()
appConfig := config.GetApp()

// 使用配置
addr := serverConfig.Addr
maxSize := imageConfig.ImageMaxSize
secret := jwtConfig.Secret
```

### 3. 生成默认配置文件

使用配置生成工具创建默认配置文件：

```bash
# 生成到当前目录
go run cmd/config-gen/main.go

# 生成到指定路径
go run cmd/config-gen/main.go -output configs/demo520.yaml

# 查看帮助
go run cmd/config-gen/main.go -help
```

## 配置文件查找顺序

1. 命令行指定的配置文件路径
2. `$HOME/.demo520/demo520.yaml`
3. 当前目录下的 `demo520.yaml`

## 环境变量支持

配置支持通过环境变量覆盖，环境变量前缀为 `DEMO520_`，配置键中的 `.` 和 `-` 会被替换为 `_`。

示例：
```bash
# 设置服务器地址
export DEMO520_SERVER_ADDR=":9090"

# 设置数据库主机
export DEMO520_DB_HOST="192.168.1.100:3306"

# 设置日志级别
export DEMO520_LOG_LEVEL="debug"
```

## 迁移指南

### 从旧的 Viper 使用方式迁移

**旧方式：**
```go
viper.GetString("server.addr")
viper.GetInt64("image.ImageMaxSize")
viper.GetBool("log.disable-caller")
```

**新方式：**
```go
config.GetServer().Addr
config.GetImage().ImageMaxSize
config.GetLog().DisableCaller
```

### 迁移步骤

1. 导入新的配置包：`import "demo520/internal/pkg/config"`
2. 在应用启动时调用 `config.Init(cfgFile)`
3. 将所有 `viper.Get*()` 调用替换为相应的配置获取方法
4. 移除不必要的 `viper.SetDefault()` 调用（已在配置包中统一设置）
5. 生成并使用新的配置文件格式

## 最佳实践

1. **集中管理**：所有配置相关的代码都在 `internal/pkg/config` 包中
2. **类型安全**：使用结构体而不是字符串键来访问配置
3. **默认值**：在配置包中统一设置所有默认值
4. **文档化**：配置文件包含详细的注释说明
5. **环境隔离**：通过环境变量支持不同环境的配置覆盖

## 故障排除

### 常见问题

1. **配置文件未找到**：检查配置文件路径和权限
2. **环境变量不生效**：确认环境变量名称格式正确
3. **配置解析失败**：检查 YAML 文件格式是否正确
4. **默认值不生效**：确认已调用 `config.Init()`

### 调试技巧

1. 启用调试日志查看配置加载过程
2. 使用 `config.GetConfig()` 打印完整配置进行验证
3. 检查环境变量是否正确设置：`env | grep DEMO520`