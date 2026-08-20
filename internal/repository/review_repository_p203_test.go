package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"gorm.io/gorm"
)

// TestReviewRepoChainErrNotFoundP203 仓储层 ErrNotFound 必须能被 errors.Is 识别（错误链不断裂）。
func TestReviewRepoChainErrNotFoundP203(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewReviewRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews" WHERE "reviews"."id" = $1 ORDER BY "reviews"."id" LIMIT $2`)).
		WithArgs(99, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByID(context.Background(), 99)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("ErrNotFound lost in chain: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestReviewRepoLockChainErrNotFoundP204 锁查询路径的 ErrNotFound 也必须能被 errors.Is 识别。
func TestReviewRepoLockChainErrNotFoundP204(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewReviewRepository(gormDB)

	mock.ExpectQuery(`SELECT .* FROM "reviews" WHERE "reviews"\."id" = \$1 ORDER BY "reviews"\."id" LIMIT \$2`).
		WithArgs(99, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByIDForUpdate(context.Background(), 99)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("ErrNotFound lost in chain: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
