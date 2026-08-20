# BUG_REPRO：终审结论 final_decision 取值体系错位，详情缺中文结论

## Bug 是什么

- `internal/service/paper_service.go`：终审 `FinalDecision` 把论文状态值 `accepted`/`rejected` 直接写进 `paper.FinalDecision`（该字段应为审稿决策值 `accept`/`reject`），且不校验非法决策值；初审拒稿 `InitialReview` 同样把 `rejected` 写进该字段。
- 整条「终审结论中文文本」链路缺失：`model.Paper` 没有 `FinalDecisionText` 字段，`Detail` 也不调用 `util.FormatReviewDecision` 渲染中文结论，前端拿到的是英文裸值。

## 如何触发

1. 终审录用（`POST /papers/:id/final-decision`，decision=accepted）后查看论文详情；
2. 初审直接拒稿（`POST /papers/:id/initial-review`，pass=false）后查看论文详情；
3. 传入非法决策值（如 banana）时，状态机被推进到未知终态。

## 真实错误信息

`go test` 实测输出（关键段）：

```
final_decision_p701_test.go:97:12: detail.FinalDecisionText undefined (type *model.Paper has no field or method FinalDecisionText)
```
