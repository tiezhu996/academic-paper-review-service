package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/service"
)

type fakeAuditRepo struct {
	repository.AuditLogRepository
	logs []model.AuditLog
}

func (f *fakeAuditRepo) Create(ctx context.Context, l *model.AuditLog) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.logs = append(f.logs, *l)
	return nil
}

type fakeAuditStore struct {
	repository.Store
	audit *fakeAuditRepo
}

func (f *fakeAuditStore) AuditLogRepository() repository.AuditLogRepository { return f.audit }

// TestAuditMiddlewareCancelledCtxP502 请求取消后审计日志仍然必须写入，不能因为 ctx 取消而丢日志。
func TestAuditMiddlewareCancelledCtxP502(t *testing.T) {
	store := &fakeAuditStore{audit: &fakeAuditRepo{}}
	auditSvc := service.NewAuditLogService(store, newTestLogger())
	mw := NewAudit(auditSvc, newTestLogger())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/write", mw.Audit(), func(c *gin.Context) {
		ctx, cancel := context.WithCancel(c.Request.Context())
		cancel()
		c.Request = c.Request.WithContext(ctx)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/write", nil))
	if len(store.audit.logs) != 1 {
		t.Fatalf("audit log missing after request ctx cancelled, logs=%d", len(store.audit.logs))
	}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
