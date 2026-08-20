# BUG_REPRO：审计/对象存储错误被吞掉，接口假成功

## Bug 是什么

- `internal/service/audit_log_service.go`：`Record` / `List` 用命名返回值 + `defer` 把写库/查询错误吞成成功（`err = nil`），审计写失败时接口仍显示成功、日志查询失败时返回空列表。
- `internal/service/storage_service.go`：`Upload` 用 `defer` 吞掉 PutObject 错误，`EnsureBucket` 吞掉建桶错误，上传失败时接口照样回成功。

## 如何触发

1. 对象存储不可用（上传/建桶返回 500）时调用文件上传，接口仍回成功；
2. 审计日志写入失败时调用任意写接口，错误被吞。

## 真实错误信息

`go test` 实测输出（关键段）：

```
storage_swallow_p602_test.go:31: expected error when object store upload fails
audit_swallow_p601_test.go:29: expected error when audit repo fails
```
