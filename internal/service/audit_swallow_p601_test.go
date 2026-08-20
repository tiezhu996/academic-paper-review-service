package service

import (
	"context"
	"errors"
	"testing"

	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
)

type failingAuditRepo struct {
	repository.AuditLogRepository
	err error
}

func (f *failingAuditRepo) Create(ctx context.Context, l *model.AuditLog) error { return f.err }

func (f *failingAuditRepo) List(ctx context.Context, page, size int) ([]model.AuditLog, int64, error) {
	return nil, 0, f.err
}

type failingStore struct {
	repository.Store
	audit *failingAuditRepo
}

func (f *failingStore) AuditLogRepository() repository.AuditLogRepository { return f.audit }

// TestAuditRecordErrorSurfacesP601 审计日志写入失败必须透出错误，不能吞成成功。
func TestAuditRecordErrorSurfacesP601(t *testing.T) {
	store := &failingStore{audit: &failingAuditRepo{err: errors.New("audit db down")}}
	svc := NewAuditLogService(store, newTestLogger())
	ctx := context.Background()

	err := svc.Record(ctx, 1, "u", "create_paper", "paper", "1", "新建投稿", "127.0.0.1", "rid-1")
	if err == nil {
		t.Fatalf("expected error when audit repo fails")
	}
}

// TestAuditListErrorSurfacesP604 审计日志查询失败必须透出错误，不能返回空列表假成功。
func TestAuditListErrorSurfacesP604(t *testing.T) {
	store := &failingStore{audit: &failingAuditRepo{err: errors.New("audit db down")}}
	svc := NewAuditLogService(store, newTestLogger())
	ctx := context.Background()

	_, _, err := svc.List(ctx, 1, 10)
	if err == nil {
		t.Fatalf("expected error when audit repo fails")
	}
}
