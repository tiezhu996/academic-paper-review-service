package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/service"
)

// TestPlagiarismHandlerCacheRaceP104 并发访问查重结果/重跑查重接口，全程不得出现 data race。
func TestPlagiarismHandlerCacheRaceP104(t *testing.T) {
	store := newFakeStore()
	logger := newTestLogger()
	cfg := newTestConfig()
	plagSvc := service.NewPlagiarismService(store, cfg, logger)
	auditSvc := service.NewAuditLogService(store, logger)
	h := NewPlagiarismHandler(plagSvc, auditSvc, logger)
	ctx := context.Background()

	paper := &model.Paper{Title: "handler-race-paper", Keywords: "h1,h2", Status: constants.PaperStatusSubmitted}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	if err := store.plag.Create(ctx, &model.PlagiarismCheck{PaperID: paper.ID, Status: constants.PlagiarismStatusPending}); err != nil {
		t.Fatalf("create check: %v", err)
	}
	if _, err := plagSvc.RunCheck(ctx, paper); err != nil {
		t.Fatalf("warm run: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/papers/:id/plagiarism", h.GetByPaper)
	r.POST("/papers/:id/plagiarism/rerun", h.Rerun)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 40; j++ {
				req := httptest.NewRequest(http.MethodGet, "/papers/1/plagiarism", nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != http.StatusOK {
					t.Errorf("get plagiarism status=%d", w.Code)
					return
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for j := 0; j < 20; j++ {
			req := httptest.NewRequest(http.MethodPost, "/papers/1/plagiarism/rerun", strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("rerun status=%d", w.Code)
				return
			}
		}
	}()
	close(start)
	wg.Wait()
}
