# BUG_REPRO：审稿任务不存在时被误判成 500

## Bug 是什么

- 审稿仓储（`internal/repository/review_repository.go`）在 `FindByID` / `FindByIDForUpdate` 里用 `%v` 包装 `ErrNotFound`，错误链断裂，`errors.Is(err, repository.ErrNotFound)` 永远为 false。
- 审稿服务（`internal/service/review_service.go`）的 `Submit` / `Respond` 因此无法识别「审稿任务不存在」，把 not-found 当作系统内部错误返回 500，而不是 404。

## 如何触发

1. 对一个不存在的审稿任务提交评审意见（`POST /reviews/:id/submit`）；
2. 或对一个不存在的审稿任务回应邀请（`POST /reviews/:id/respond`）。

## 真实错误信息

`go test` 实测输出（关键行）：

```
code = 50000, want 40403 (err=审稿回应失败：系统内部错误: record not found)
```

接口层面表现为回 500（内部错误码 50000），应为 404（业务码 40403 / ErrReviewNotFound）。
