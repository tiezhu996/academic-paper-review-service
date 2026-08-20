# BUG_REPRO：查重结果缓存并发读写导致进程崩溃与快照污染

## Bug 是什么

- 查重服务（`internal/service/plagiarism_service.go`）与查重 HTTP 处理器（`internal/handler/plagiarism_handler.go`）各自维护一个进程内的「最近查重结果缓存」（`recent map[uint]*model.PlagiarismCheck`）。
- 缓存读写没有加锁：`RunCheck`/`GetByPaper`/`Rerun` 并发访问同一个 map；`RunCheck` 在重跑时还会直接复用并就地改写缓存中旧对象的字段。
- 后果一：并发读写触发 `fatal error: concurrent map read and map write` 或 `-race` 报 `WARNING: DATA RACE`，整个服务进程崩溃。
- 后果二：之前返回给调用方的查重快照被后续重跑原地改写，出现新旧结果串数据。

## 如何触发

1. 编辑重跑某篇论文的查重（`POST /papers/:id/plagiarism/rerun`）的同时，反复查看该论文查重结果（`GET /papers/:id/plagiarism`）；
2. 或作者投稿触发自动查重（`POST /papers`）的同时，多次查看查重结果。

并发点几下即可复现。

## 真实错误信息

`go test -race` 实测输出（关键段）：

```
WARNING: DATA RACE
Write at 0x00c0003fe870 by goroutine 9:
	internal/service/plagiarism_service.go:61 +0x6b0
	internal/service/plagiarism_cache_p101_test.go:37 +0xd4
Previous read at 0x00c0003fe870 by goroutine 11:
	internal/service/plagiarism_service.go:77 +0x58
```

无 `-race` 时表现为 `fatal error: concurrent map read and map write`，服务进程直接退出。
