# BUG_REPRO：请求取消后仍继续写库（ctx 取消未传播）

## Bug 是什么

- `internal/repository/user_repository.go`：`Create` / `FindByID` / `FindByUsername` 方法签名接收 `ctx`，函数体却用 `context.Background()` 执行 SQL，请求的取消信号传不到数据库驱动层。
- `internal/service/auth_service.go`：`Register` / `Login` 开头不检查 `ctx.Err()`，请求已取消仍继续创建用户 / 签发令牌。
- `internal/middleware/audit.go`：请求结束后审计写入使用已取消的 request ctx，审计日志可能被丢弃。

## 如何触发

1. 注册 / 登录请求超时或客户端取消后，后台仍把用户建出来、仍签发 token；
2. 请求取消后审计日志写入失败。

## 真实错误信息

`go test` 实测输出（关键段）：

```
auth_ctx_p501_test.go:24: expected error when ctx already cancelled
audit_ctx_p502_test.go:55: audit log missing after request ctx cancelled, logs=0
```
