package service

import (
	"context"
	"testing"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/model"
)

func newPaperTestEnvForDecision(t *testing.T) (*fakeStore, *PaperService) {
	t.Helper()
	store := newFakeStore()
	cfg := newTestConfig()
	plagiarismSvc := NewPlagiarismService(store, cfg, newTestLogger())
	svc := NewPaperService(store, plagiarismSvc, cfg, newTestLogger())
	return store, svc
}

// TestFinalDecisionValueSystemP701 终审录用后 final_decision 必须使用审稿决策取值体系（accept），不是论文状态值。
func TestFinalDecisionValueSystemP701(t *testing.T) {
	store, svc := newPaperTestEnvForDecision(t)
	ctx := context.Background()
	paper := &model.Paper{Title: "P", Status: constants.PaperStatusExternalReview, SubmitterID: 1}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}

	updated, err := svc.FinalDecision(ctx, 2, paper.ID, dto.FinalDecisionRequest{
		Decision: constants.PaperStatusAccepted, Comment: "同意录用",
	})
	if err != nil {
		t.Fatalf("final decision: %v", err)
	}
	if updated.Status != constants.PaperStatusAccepted {
		t.Fatalf("status = %s, want accepted", updated.Status)
	}
	if updated.FinalDecision != constants.ReviewDecisionAccept {
		t.Fatalf("final_decision = %q, want %q", updated.FinalDecision, constants.ReviewDecisionAccept)
	}
}

// TestFinalDecisionInvalidRejectedP702 终审非法决策值必须被拒绝，不能把状态机推进到非法终态。
func TestFinalDecisionInvalidRejectedP702(t *testing.T) {
	store, svc := newPaperTestEnvForDecision(t)
	ctx := context.Background()
	paper := &model.Paper{Title: "P", Status: constants.PaperStatusExternalReview, SubmitterID: 1}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}

	_, err := svc.FinalDecision(ctx, 2, paper.ID, dto.FinalDecisionRequest{
		Decision: "banana", Comment: "x",
	})
	if err == nil {
		t.Fatalf("expected error for invalid decision")
	}
}

// TestInitialReviewRejectDecisionValueP703 初审直接拒稿后 final_decision 必须使用审稿决策取值体系（reject）。
func TestInitialReviewRejectDecisionValueP703(t *testing.T) {
	store, svc := newPaperTestEnvForDecision(t)
	ctx := context.Background()
	paper := &model.Paper{Title: "P", Status: constants.PaperStatusSubmitted, SubmitterID: 1}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}

	updated, err := svc.InitialReview(ctx, 2, paper.ID, dto.InitialReviewRequest{Pass: false, Reason: "选题不符"})
	if err != nil {
		t.Fatalf("initial review: %v", err)
	}
	if updated.FinalDecision != constants.ReviewDecisionReject {
		t.Fatalf("final_decision = %q, want %q", updated.FinalDecision, constants.ReviewDecisionReject)
	}
}

// TestPaperDetailFinalDecisionTextP704 终审录用后详情必须带中文结论文本。
func TestPaperDetailFinalDecisionTextP704(t *testing.T) {
	store, svc := newPaperTestEnvForDecision(t)
	ctx := context.Background()
	paper := &model.Paper{Title: "P", Status: constants.PaperStatusExternalReview, SubmitterID: 1}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	if _, err := svc.FinalDecision(ctx, 2, paper.ID, dto.FinalDecisionRequest{
		Decision: constants.PaperStatusAccepted, Comment: "同意录用",
	}); err != nil {
		t.Fatalf("final decision: %v", err)
	}

	detail, err := svc.Detail(ctx, paper.ID)
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if detail.FinalDecisionText != "录用" {
		t.Fatalf("final_decision_text = %q, want 录用", detail.FinalDecisionText)
	}
}
