# 日志最佳实践指南

本文档提供了在项目中使用日志系统的最佳实践和示例。

## 日志记录原则

### 1. 安全性原则
- **永远不要记录敏感信息**：密码、JWT令牌、个人身份信息等
- **使用占位符**：对于敏感字段使用 `[REDACTED]` 或 `***`
- **记录存在性而非内容**：使用 `has_password: true/false` 而不是密码内容

### 2. 可观测性原则
- **记录关键业务流程**：用户登录、注册、密码修改等
- **记录性能指标**：数据库查询时间、API响应时间
- **记录错误上下文**：足够的信息用于问题诊断

### 3. 一致性原则
- **统一的字段命名**：使用一致的字段名称（如 `user_uuid`, `email`, `client_ip`）
- **统一的日志级别**：按照既定规则使用不同的日志级别
- **统一的消息格式**：使用清晰、一致的日志消息

## 控制器层日志示例

### 用户登录控制器
```go
func (ctrl *UserController) Login(c *gin.Context) {
    defer log.C(c).FuncEntryWithContext(c, "email", "[REDACTED]")() // 函数进入/退出追踪

    log.C(c).Infow("User login attempt started",
        "client_ip", c.ClientIP(),
        "user_agent", c.Request.UserAgent(),
    )

    var r api.LoginRequest
    if err := c.ShouldBindJSON(&r); err != nil {
        log.C(c).Errorw("Failed to bind login request",
            "error", err.Error(),
            "client_ip", c.ClientIP(),
        )
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    log.C(c).Infow("Login request validation started",
        "email", r.Email, // 邮箱不是敏感信息
        "has_password", r.Password != "",
        "seed_time", r.SeedTime,
    )

    // ... 业务逻辑处理

    log.C(c).Infow("Login successful",
        "email", r.Email,
        "client_ip", c.ClientIP(),
        "has_token", resp.Token != "",
    )
}
```

### 密码修改控制器
```go
func (ctrl *UserController) ChangePassword(c *gin.Context) {
    defer log.C(c).FuncEntryWithContext(c, "email", "[REDACTED]")() 

    email := c.Param("email")
    
    log.C(c).Infow("Password change attempt started",
        "email", email,
        "client_ip", c.ClientIP(),
        "user_agent", c.Request.UserAgent(),
    )

    // 记录密码字段的存在性，而不是内容
    log.C(c).Infow("Password change request validation started",
        "email", email,
        "has_old_password", r.OldPassword != "",
        "has_new_password", r.NewPassword != "",
    )

    // ... 处理逻辑
}
```

## 业务层日志示例

### 用户业务逻辑
```go
func (u *userBiz) Login(ctx context.Context, r *api.LoginRequest) (*api.LoginResponse, error) {
    defer log.C(ctx).FuncEntryWithContext(ctx, r.Email, "***", r.SeedTime)()

    log.C(ctx).Infow("Business layer login started",
        "email", r.Email,
        "seed_time", r.SeedTime,
    )

    // 时间戳解析
    log.C(ctx).Debugw("Parsing seed time", "seed_time", r.SeedTime)
    
    // 时间验证
    log.C(ctx).Debugw("Validating request timing",
        "current_time", now,
        "request_time", utcTime,
        "duration", duration.String(),
        "timeout", timeout.String(),
    )

    // 用户查询
    log.C(ctx).Infow("Fetching user from database", "email", r.Email)
    
    // 密码验证
    log.C(ctx).Infow("User found, verifying password",
        "email", r.Email,
        "user_uuid", userM.UserUUID,
        "created_at", userM.CreatedAt,
    )

    if !auth.VerifyPassword(r.Password, userM.Password) {
        log.C(ctx).Warnw("Password verification failed",
            "email", r.Email,
            "user_uuid", userM.UserUUID,
        )
        return nil, errno.ErrPasswordIncorrect
    }

    // JWT生成
    log.C(ctx).Infow("Password verification successful, generating JWT token",
        "email", r.Email,
        "user_uuid", userM.UserUUID,
    )
}
```

## 数据存储层日志示例

### 数据库操作
```go
func (u *userStore) Get(ctx context.Context, email string) (*model.UserM, error) {
    log.C(ctx).Infow("Database query: Get user by email",
        "email", email,
        "operation", "SELECT",
        "table", "users",
    )

    start := time.Now()
    if err := u.db.Where("email = ?", email).First(&user).Error; err != nil {
        duration := time.Since(start)
        log.C(ctx).Errorw("Database query failed",
            "email", email,
            "error", err.Error(),
            "operation", "SELECT",
            "table", "users",
            "duration_ms", duration.Milliseconds(),
            "sql_error", errors.Is(err, gorm.ErrRecordNotFound),
        )
        return nil, err
    }

    duration := time.Since(start)
    log.C(ctx).Infow("Database query successful",
        "email", email,
        "user_uuid", user.UserUUID,
        "operation", "SELECT",
        "table", "users",
        "duration_ms", duration.Milliseconds(),
    )

    return &user, nil
}
```

### 事务操作
```go
func (u *userStore) ChangePassword(ctx context.Context, userUUID, oldPassword, newPassword string) error {
    log.C(ctx).Infow("Starting database transaction for password change",
        "user_uuid", userUUID,
    )

    start := time.Now()
    err := u.db.Transaction(func(tx *gorm.DB) error {
        log.C(ctx).Debugw("Fetching user for password verification",
            "user_uuid", userUUID,
        )

        // ... 事务逻辑

        log.C(ctx).Infow("Password updated successfully in transaction",
            "user_uuid", userUUID,
            "email", user.Email,
        )
        return nil
    })

    duration := time.Since(start)
    if err != nil {
        log.C(ctx).Errorw("Database transaction failed",
            "user_uuid", userUUID,
            "error", err.Error(),
            "operation", "UPDATE",
            "table", "users",
            "duration_ms", duration.Milliseconds(),
        )
        return err
    }

    log.C(ctx).Infow("Database transaction completed successfully",
        "user_uuid", userUUID,
        "operation", "UPDATE",
        "table", "users",
        "duration_ms", duration.Milliseconds(),
    )

    return nil
}
```

## 中间件日志示例

### JWT认证中间件
```go
func JWTAuth() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        log.C(ctx).Infow("JWT authentication started",
            "path", ctx.Request.URL.Path,
            "method", ctx.Request.Method,
            "client_ip", ctx.ClientIP(),
            "user_agent", ctx.Request.UserAgent(),
        )

        jwtUserUUID, err := token.ParseRequest(ctx)
        if err != nil {
            if errors.Is(err, token.ErrMissingHeader) {
                log.C(ctx).Infow("JWT token missing, allowing anonymous access",
                    "path", ctx.Request.URL.Path,
                    "method", ctx.Request.Method,
                    "client_ip", ctx.ClientIP(),
                )
            } else {
                log.C(ctx).Warnw("JWT authentication failed",
                    "error", err.Error(),
                    "path", ctx.Request.URL.Path,
                    "method", ctx.Request.Method,
                    "client_ip", ctx.ClientIP(),
                )
            }
        } else {
            log.C(ctx).Infow("JWT authentication successful",
                "user_uuid", jwtUserUUID,
                "path", ctx.Request.URL.Path,
                "method", ctx.Request.Method,
                "client_ip", ctx.ClientIP(),
            )
        }
    }
}
```

### HTTP请求追踪中间件
```go
func RequestTracing() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        requestID := generateRequestID()
        
        // 设置请求ID到上下文
        c.Set(known.XRequestIDKey, requestID)
        
        log.C(c).Infow("HTTP request started",
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
            "query", c.Request.URL.RawQuery,
            "client_ip", c.ClientIP(),
            "user_agent", c.Request.UserAgent(),
        )

        // 处理请求
        c.Next()

        // 记录响应
        duration := time.Since(start)
        log.C(c).Infow("HTTP request completed",
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
            "status_code", c.Writer.Status(),
            "duration_ms", duration.Milliseconds(),
            "response_size", c.Writer.Size(),
        )
    }
}
```

## 常用日志字段

### 标准字段
- `request_id`: 请求唯一标识
- `user_uuid`: 用户唯一标识
- `email`: 用户邮箱
- `client_ip`: 客户端IP地址
- `user_agent`: 用户代理字符串
- `method`: HTTP方法
- `path`: 请求路径
- `status_code`: HTTP状态码
- `duration_ms`: 操作耗时（毫秒）

### 数据库字段
- `operation`: 数据库操作类型（SELECT/INSERT/UPDATE/DELETE）
- `table`: 操作的表名
- `rows_affected`: 影响的行数
- `sql`: SQL语句（生产环境可选）
- `sql_error`: 是否为SQL错误

### 业务字段
- `has_password`: 是否包含密码
- `has_token`: 是否包含令牌
- `operation_type`: 业务操作类型
- `result`: 操作结果

## 日志级别使用指南

### Debug级别
- 详细的执行流程
- 参数解析过程
- 中间计算结果
- 仅在开发环境使用

### Info级别
- 正常的业务操作
- 成功的数据库查询
- 用户行为记录
- 系统状态变化

### Warn级别
- 可恢复的错误
- 性能问题（如慢查询）
- 安全相关事件（如登录失败）
- 配置问题

### Error级别
- 不可恢复的错误
- 数据库连接失败
- 外部服务调用失败
- 系统异常

## 性能考虑

1. **避免过度日志记录**：在高频操作中谨慎使用Debug级别
2. **异步日志写入**：考虑使用异步日志以减少性能影响
3. **日志轮转**：配置适当的日志轮转策略
4. **结构化日志**：使用JSON格式便于日志分析工具处理

## 监控和告警

基于日志建立监控和告警机制：

1. **错误率监控**：监控Error级别日志的频率
2. **性能监控**：监控`duration_ms`字段识别性能问题
3. **安全监控**：监控登录失败、权限错误等安全事件
4. **业务监控**：监控关键业务操作的成功率