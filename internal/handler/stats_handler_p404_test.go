package handler

import (
	"context"
	"io"
	"log/slog"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/service"
)

type fakeReviewRepoForStats struct {
	repository.ReviewRepository
	loadRows []model.ReviewerLoad
}

func (f *fakeReviewRepoForStats) CountCompletedByReviewer(ctx context.Context) ([]model.ReviewerLoad, error) {
	return f.loadRows, nil
}

type fakeStoreForStats struct {
	repository.Store
	reviews *fakeReviewRepoForStats
}

func (f *fakeStoreForStats) ReviewRepository() repository.ReviewRepository { return f.reviews }
func (f *fakeStoreForStats) Transaction(ctx context.Context, fn func(repository.Store) error) error {
	return fn(f)
}

// TestStatsReviewerWorkloadFreshP404 审稿人工作量数据变化后，再次请求必须返回最新数据，不能返回过期缓存。
func TestStatsReviewerWorkloadFreshP404(t *testing.T) {
	store := &fakeStoreForStats{reviews: &fakeReviewRepoForStats{}}
	statSvc := service.NewStatisticsService(store, newTestLogger())
	h := NewStatisticsHandler(statSvc, newTestLogger())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/statistics/reviewers", h.ReviewerWorkload)

	store.reviews.loadRows = []model.ReviewerLoad{{ReviewerID: 1, ReviewerName: "审稿人甲", Total: 5, Completed: 3}}
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/statistics/reviewers", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("first status = %d", w1.Code)
	}

	store.reviews.loadRows = []model.ReviewerLoad{{ReviewerID: 2, ReviewerName: "审稿人乙", Total: 9, Completed: 4}}
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/statistics/reviewers", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("second status = %d", w2.Code)
	}
	var resp struct {
		Data []*model.ReviewerLoad `json:"data"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, w2.Body.String())
	}
	if len(resp.Data) != 1 || resp.Data[0].ReviewerID != 2 {
		t.Fatalf("second request returned stale data: %+v", resp.Data)
	}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
