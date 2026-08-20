# BUG_REPRO：修稿记录列表缓存并发读写导致 data race

## Bug 是什么

- `internal/service/revision_service.go`：`ListByPaper` 维护单槽缓存（`lastPaperID`/`lastItems`），并发查询时无锁读写缓存字段（data race）；查询另一篇论文时用 `append(lastItems[:0], ...)` 复用底层数组，把第一次返回给调用方的列表原地改写；命中缓存时直接返回内部切片引用。
- `internal/repository/revision_repository.go`：`ListByPaper` 复用结构体缓冲切片（`r.buf[:0]`），连续查询第一次的结果被第二次改写。

## 如何触发

1. 多个请求并发查询同一/不同论文的修稿记录（`GET /papers/:id/revisions`）；
2. 连续查询不同论文，对比第一次返回的列表是否被第二次改写。

## 真实错误信息

`go test -race` 实测输出（关键段）：

```
WARNING: DATA RACE
Write at 0x00c000325d98 by goroutine 9:
	internal/service/revision_service.go:37 +0x178
	internal/service/revision_cache_p901_test.go:29 +0xcc
```
