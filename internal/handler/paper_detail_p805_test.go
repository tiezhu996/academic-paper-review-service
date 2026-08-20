package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/config"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/service"
)

type fakePaperRepoForDetail struct {
	repository.PaperRepository
}

func (f *fakePaperRepoForDetail) FindByIDWithDetail(ctx context.Context, id uint) (*model.Paper, error) {
	return nil, repository.ErrNotFound
}

type fakeStoreForDetail struct {
	repository.Store
	paper *fakePaperRepoForDetail
}

func (f *fakeStoreForDetail) PaperRepository() repository.PaperRepository { return f.paper }
func (f *fakeStoreForDetail) Transaction(ctx context.Context, fn func(repository.Store) error) error {
	return fn(f)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestConfig() *config.Config {
	return &config.Config{JWTSecret: "test-secret", JWTExpireHours: 72, SimilarityThreshold: 30}
}

// TestPaperDetailMissingHTTPStatusP805 查询不存在的论文详情必须回 404，不能误判成 500。
func TestPaperDetailMissingHTTPStatusP805(t *testing.T) {
	store := &fakeStoreForDetail{paper: &fakePaperRepoForDetail{}}
	logger := newTestLogger()
	cfg := newTestConfig()
	plagSvc := service.NewPlagiarismService(store, cfg, logger)
	paperSvc := service.NewPaperService(store, plagSvc, cfg, logger)
	auditSvc := service.NewAuditLogService(store, logger)
	h := NewPaperHandler(paperSvc, auditSvc, logger)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/papers/:id", h.Detail)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/papers/999", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}
