# BUG_REPRO：论文/查重记录不存在时被误判成 500

## Bug 是什么

- `internal/repository/paper_repository.go`：`FindByID` / `FindByIDWithDetail` / `FindByIDForUpdate` 用 `%v` 包装 `ErrNotFound`，错误链断裂，`errors.Is(err, repository.ErrNotFound)` 失效。
- `internal/repository/plagiarism_repository.go`：`FindByPaper` 同样 `%v` 断链。
- `internal/handler/paper_handler.go`：`wrapError` 把所有业务错误一律映射成 500。
- 结果：查询/操作不存在的论文或查重记录返回 500（应 404）。

## 如何触发

1. `GET /papers/:id`、`PUT /papers/:id`、`POST /papers/:id/initial-review` 等对不存在的论文；
2. `GET /papers/:id/plagiarism`、`POST /papers/:id/plagiarism/rerun` 对不存在的查重记录。

## 真实错误信息

`go test` 实测输出（关键段）：

```
code = 50000, want 40403 (err=审稿回应失败：系统内部错误: record not found)
```
（handler 层表现为不存在的论文详情回 500 而不是 404。）
