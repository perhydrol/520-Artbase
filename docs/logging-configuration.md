# 日志配置指南

本文档描述了如何配置项目的日志系统以获得最佳的可观测性。

## 日志级别配置

### 开发环境推荐配置

```yaml
# config/dev.yaml
log:
  level: debug
  format: console
  output_paths:
    - stdout
    - logs/app.log
  error_output_paths:
    - stderr
    - logs/error.log

database:
  mysql:
    log_level: 4  # Info级别，记录所有SQL查询
```

### 生产环境推荐配置

```yaml
# config/prod.yaml
log:
  level: info
  format: json
  output_paths:
    - logs/app.log
  error_output_paths:
    - logs/error.log

database:
  mysql:
    log_level: 2  # Warn级别，只记录慢查询和错误
```

## 日志级别说明

### 应用日志级别
- `debug`: 详细的调试信息，包括函数进入/退出、参数验证等
- `info`: 一般信息，包括请求处理、业务操作成功等
- `warn`: 警告信息，如密码验证失败、慢查询等
- `error`: 错误信息，如数据库操作失败、业务逻辑错误等

### 数据库日志级别
- `1` (Silent): 不记录任何SQL日志
- `2` (Error): 只记录SQL错误
- `3` (Warn): 记录SQL错误和慢查询
- `4` (Info): 记录所有SQL查询

## 关键日志字段

### HTTP请求日志
- `request_id`: 请求唯一标识
- `method`: HTTP方法
- `path`: 请求路径
- `client_ip`: 客户端IP
- `user_agent`: 用户代理
- `duration_ms`: 请求处理时间
- `status_code`: 响应状态码

### 业务操作日志
- `user_uuid`: 用户唯一标识
- `email`: 用户邮箱（非敏感信息）
- `operation`: 操作类型
- `result`: 操作结果

### 数据库操作日志
- `sql`: SQL语句
- `duration_ms`: 执行时间
- `rows_affected`: 影响行数
- `operation`: 操作类型（SELECT/INSERT/UPDATE/DELETE）
- `table`: 操作的表名

## 敏感信息处理

### 不应记录的敏感信息
- 用户密码（明文或加密后）
- JWT令牌完整内容
- 个人身份信息（如身份证号）
- 支付相关信息

### 安全记录方式
- 密码字段使用 `[REDACTED]` 或 `***` 替代
- JWT令牌只记录长度或前几位字符
- 邮箱地址可以记录（通常不被视为敏感信息）
- 使用 `has_password: true/false` 代替记录密码内容

## 日志查询示例

### 查找特定用户的登录记录
```bash
# 使用jq查询JSON格式日志
cat logs/app.log | jq 'select(.email == "user@example.com" and .msg == "Login successful")'
```

### 查找慢查询
```bash
# 查找执行时间超过500ms的数据库查询
cat logs/app.log | jq 'select(.duration_ms > 500 and .sql != null)'
```

### 查找错误日志
```bash
# 查找所有错误级别的日志
cat logs/app.log | jq 'select(.level == "error")'
```

## 监控和告警建议

### 关键指标监控
1. **错误率**: 监控HTTP 5xx错误和数据库错误
2. **响应时间**: 监控API响应时间和数据库查询时间
3. **登录失败率**: 监控密码验证失败的频率
4. **慢查询**: 监控执行时间超过阈值的SQL查询

### 告警规则示例
- HTTP错误率超过5%
- 平均响应时间超过1秒
- 数据库连接失败
- 连续登录失败超过5次（可能的暴力破解）

## 日志轮转配置

建议使用logrotate或类似工具进行日志轮转：

```bash
# /etc/logrotate.d/demo520
/path/to/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0644 app app
    postrotate
        # 发送信号给应用重新打开日志文件
        killall -USR1 demo520 || true
    endscript
}
```

## 性能考虑

1. **生产环境**: 使用JSON格式便于日志分析工具处理
2. **开发环境**: 使用console格式便于人工阅读
3. **异步日志**: 考虑使用异步日志写入以减少对性能的影响
4. **日志级别**: 生产环境避免使用debug级别以减少日志量