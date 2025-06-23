# 日志系统使用指南

本文档展示了如何在项目中使用增强的日志系统，包括函数入口跟踪、参数记录和错误处理。

## 新增功能

### 1. 函数入口跟踪

#### 基本用法
```go
func SomeFunction(param1 string, param2 int) error {
    defer log.FuncEntry(param1, param2)()
    
    // 函数逻辑
    return nil
}
```

#### 带上下文的用法
```go
func SomeFunctionWithContext(ctx context.Context, param1 string, param2 int) error {
    defer log.FuncEntryWithContext(ctx, param1, param2)()
    
    // 函数逻辑
    return nil
}
```

### 2. 错误处理增强

#### 使用 ErrorWithFunc 记录错误
```go
func SomeFunction() error {
    defer log.FuncEntry()()
    
    if err := someOperation(); err != nil {
        log.ErrorWithFunc(err, "操作失败", "operation", "someOperation")
        return err
    }
    
    return nil
}
```

## 实际应用示例

### 业务层函数示例

```go
func (i *imageBiz) Create(ctx context.Context, userUUID string, r *api.CreateImageRequest, fileHeader *multipart.FileHeader) (*api.CreateImageResponse, error) {
    defer log.FuncEntryWithContext(ctx, userUUID, r, fileHeader)()
    
    // 参数验证
    if fileHeader == nil {
        err := errno.ErrInvalidParameter.SetMessage("file header is required")
        log.ErrorWithFunc(err, "参数验证失败", "parameter", "fileHeader")
        return nil, err
    }
    
    if r == nil {
        err := errno.ErrInvalidParameter.SetMessage("request is required")
        log.ErrorWithFunc(err, "参数验证失败", "parameter", "request")
        return nil, err
    }
    
    // 文件大小验证
    imageMaxSize := config.GetImage().ImageMaxSize
    if fileHeader.Size > imageMaxSize {
        err := errno.ErrImageFileTooLarge
        log.ErrorWithFunc(err, "文件大小超限", 
            "fileSize", fileHeader.Size, 
            "maxSize", imageMaxSize)
        return nil, err
    }
    
    // 文件格式验证
    if ok, err := i.imageFileStore.Validate(fileHeader); err != nil {
        log.ErrorWithFunc(err, "文件格式验证失败")
        return nil, errno.InternalServerError.SetMessage("failed to validate image: %v", err)
    } else if !ok {
        err := errno.ErrImageFileInvalid
        log.ErrorWithFunc(err, "文件格式无效")
        return nil, err
    }
    
    // 计算文件哈希
    hash, err := i.imageFileStore.Hash(fileHeader)
    if err != nil {
        log.ErrorWithFunc(err, "计算文件哈希失败")
        return nil, errno.InternalServerError.SetMessage("failed to calculate image hash: %v", err)
    }
    
    log.Infow("文件哈希计算成功", "hash", hash)
    
    // 保存文件
    imageUUID := uuid.New()
    if err := i.imageFileStore.Save(fileHeader, hash); err != nil {
        log.ErrorWithFunc(err, "保存文件失败", "imageUUID", imageUUID.String())
        return nil, errno.InternalServerError.SetMessage("failed to save image file: %v", err)
    }
    
    log.Infow("文件保存成功", "imageUUID", imageUUID.String(), "hash", hash)
    
    // 其他业务逻辑...
    
    return &api.CreateImageResponse{
        ImageUUID: imageUUID.String(),
    }, nil
}
```

### 存储层函数示例

```go
func (u *userStore) ChangePassword(ctx context.Context, email string, oldPassword string, newPassword string) error {
    defer log.FuncEntryWithContext(ctx, email, "***", "***")()
    
    return u.db.Transaction(func(tx *gorm.DB) error {
        var user model.UserM
        
        // 查找用户
        if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                log.ErrorWithFunc(err, "用户不存在", "email", email)
                return errno.ErrUserNotFound
            }
            log.ErrorWithFunc(err, "查询用户失败", "email", email)
            return err
        }
        
        // 验证旧密码
        if !helper.CheckPasswordHash(oldPassword, user.Password) {
            err := errno.ErrPasswordIncorrect
            log.ErrorWithFunc(err, "旧密码验证失败", "email", email)
            return err
        }
        
        // 加密新密码
        hashedPassword, err := helper.HashPassword(newPassword)
        if err != nil {
            log.ErrorWithFunc(err, "密码加密失败")
            return err
        }
        
        // 更新密码
        if err := tx.Model(&user).Update("password", hashedPassword).Error; err != nil {
            log.ErrorWithFunc(err, "更新密码失败", "email", email)
            return err
        }
        
        log.Infow("密码更新成功", "email", email)
        return nil
    })
}
```

### 控制器层函数示例

```go
func (ctrl *UserController) ChangePassword(c *gin.Context) {
    defer log.FuncEntryWithContext(c.Request.Context())()
    
    var r api.ChangePasswordRequest
    if err := c.ShouldBindJSON(&r); err != nil {
        log.ErrorWithFunc(err, "请求参数绑定失败")
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }
    
    // 从JWT中获取用户邮箱
    email, exists := c.Get(known.XUsernameKey)
    if !exists {
        err := errno.ErrTokenInvalid
        log.ErrorWithFunc(err, "无法从JWT中获取用户信息")
        core.WriteResponse(c, err, nil)
        return
    }
    
    if err := ctrl.b.Users().ChangePassword(c.Request.Context(), email.(string), &r); err != nil {
        log.ErrorWithFunc(err, "修改密码失败", "email", email)
        core.WriteResponse(c, err, nil)
        return
    }
    
    log.Infow("密码修改成功", "email", email)
    core.WriteResponse(c, nil, nil)
}
```

## 日志输出示例

使用新的日志系统后，你将看到如下格式的日志输出：

```
2024-01-15T10:30:45.123Z DEBUG Function entry {"function": "Create", "action": "entry", "arg0": "user-uuid-123", "arg1": "api.CreateImageRequest{...}", "arg2": "&multipart.FileHeader{...}", "request_id": "req-123", "username": "user@example.com"}

2024-01-15T10:30:45.125Z INFO 文件哈希计算成功 {"hash": "abc123def456", "request_id": "req-123", "username": "user@example.com"}

2024-01-15T10:30:45.130Z INFO 文件保存成功 {"imageUUID": "img-uuid-456", "hash": "abc123def456", "request_id": "req-123", "username": "user@example.com"}

2024-01-15T10:30:45.135Z DEBUG Function exit {"function": "Create", "action": "exit", "duration": "12.345ms", "request_id": "req-123", "username": "user@example.com"}
```

## 最佳实践

1. **函数入口记录**：在所有重要的业务函数开头使用 `defer log.FuncEntry()()` 或 `defer log.FuncEntryWithContext(ctx)()`

2. **参数记录**：记录重要参数，但注意敏感信息（如密码）应该用占位符替代

3. **错误处理**：使用 `log.ErrorWithFunc()` 记录错误，提供足够的上下文信息

4. **关键操作记录**：在重要操作完成后使用 `log.Infow()` 记录成功信息

5. **性能监控**：函数入口跟踪会自动记录函数执行时间，有助于性能分析

6. **上下文传递**：在有 context 的函数中优先使用 `FuncEntryWithContext`，这样可以自动关联请求ID和用户信息