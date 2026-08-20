# BUG_REPRO：统计接口连续查询时快照被改写 / 返回过期缓存

## Bug 是什么

- `internal/service/statistics_service.go`：`Trend` / `Subjects` / `ReviewerWorkload` 把最近一次结果写进结构体共享切片（`s.trendBuf` / `s.subjectBuf` / `s.loadBuf`），用 `append(buf[:0], rows...)` 复用底层数组，并把内部切片直接返回给调用方。
- 连续两次查询时，第二次请求把新数据写进第一次调用方仍持有的同一底层数组，第一次拿到的结果被原地改写（快照污染）。
- `internal/handler/statistics_handler.go`：`ReviewerWorkload` 把最近一次响应缓存到 `h.lastWorkload` 且永不失效，数据更新后再请求仍返回旧数据。

## 如何触发

1. 连续两次请求 `GET /statistics/trend` / `/statistics/subjects` / `/statistics/reviewers`，对比第一次返回的内容是否被第二次改写；
2. 审稿人工作量数据变化后再次请求 `GET /statistics/reviewers`，观察是否返回过期缓存。

## 真实错误信息

`go test` 实测输出（关键段）：

```
first snapshot polluted: [{ReviewerID:2 ReviewerName:乙 Total:8 Completed:1}]
```
