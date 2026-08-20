# BUG_REPRO：审计/修稿列表缓冲复用污染 + 审计 IP 就地打码 + 审稿人列表过期缓存

## Bug 是什么

- `internal/repository/audit_log_repository.go`：`List` 复用结构体缓冲切片（`r.buf[:0]`），连续查询第一次的结果被第二次改写；且 `total` 返回的是当前页条数（`len(r.buf)`）而不是真实总数。
- `internal/handler/audit_handler.go`：`List` 对服务返回的共享切片就地做 IP 打码，把共享数据改掉。
- `internal/handler/user_handler.go`：`ListReviewers` 缓存最近一次审稿人列表且永不失效，数据变化后仍返回过期数据。

## 如何触发

1. 连续两次请求 `GET /audit-logs`，对比第一次返回内容与 total 是否被第二次改写/变成当前页条数；
2. 请求 `GET /audit-logs` 观察 IP 是否被就地打码；
3. 审稿人数据变化后再次请求 `GET /users/reviewers`，观察是否返回过期列表。

## 真实错误信息

`go test` 实测输出（关键段）：

```
audit_repo_p1001_test.go:59: first list polluted: [{ID:3 Action:login IP:3.3.3.3}]
audit_repo_p1001_test.go:73: total = 10, want 25
```
