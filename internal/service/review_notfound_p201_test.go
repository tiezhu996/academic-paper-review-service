package service

import (
	"context"
	"errors"
	"testing"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/util"
)

func assertAppCode(t *testing.T, err error, wantCode int) {
	t.Helper()
	var appErr *util.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %v", err)
	}
	if appErr.Code != wantCode {
		t.Fatalf("code = %d, want %d (err=%v)", appErr.Code, wantCode, err)
	}
}

// TestReviewSubmitMissingReviewP201 对不存在的审稿任务提交意见必须回「任务不存在」，而不是系统错误。
func TestReviewSubmitMissingReviewP201(t *testing.T) {
	store := newFakeStore()
	svc := NewReviewService(store, newTestLogger())
	ctx := context.Background()

	err := svc.Submit(ctx, 999, 1, dto.SubmitReviewRequest{
		Decision: constants.ReviewDecisionAccept,
		Comments: "补充实验后可以录用。",
	})
	assertAppCode(t, err, constants.ErrReviewNotFound)
}

// TestReviewRespondMissingReviewP202 对不存在的审稿任务回应邀请必须回「任务不存在」，而不是系统错误。
func TestReviewRespondMissingReviewP202(t *testing.T) {
	store := newFakeStore()
	svc := NewReviewService(store, newTestLogger())
	ctx := context.Background()

	err := svc.Respond(ctx, 999, 1, true)
	assertAppCode(t, err, constants.ErrReviewNotFound)
}
