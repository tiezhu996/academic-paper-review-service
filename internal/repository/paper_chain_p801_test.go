package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"gorm.io/gorm"
)

func assertErrNotFoundInChain(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("ErrNotFound lost in chain: %v", err)
	}
}

// TestPaperRepoChainErrNotFoundP801 FindByID 的 ErrNotFound 必须能被 errors.Is 识别。
func TestPaperRepoChainErrNotFoundP801(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewPaperRepository(gormDB)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "papers" WHERE "papers"."id" = $1 ORDER BY "papers"."id" LIMIT $2`)).
		WithArgs(99, 1).WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByID(context.Background(), 99)
	assertErrNotFoundInChain(t, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestPaperRepoFindByIDForUpdateChainP802 FindByIDForUpdate 的 ErrNotFound 必须能被 errors.Is 识别。
func TestPaperRepoFindByIDForUpdateChainP802(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewPaperRepository(gormDB)
	mock.ExpectQuery(`SELECT .* FROM "papers" WHERE "papers"\."id" = \$1 ORDER BY "papers"\."id" LIMIT \$2`).
		WithArgs(99, 1).WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByIDForUpdate(context.Background(), 99)
	assertErrNotFoundInChain(t, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestPaperRepoDetailChainErrNotFoundP804 FindByIDWithDetail 的 ErrNotFound 必须能被 errors.Is 识别。
func TestPaperRepoDetailChainErrNotFoundP804(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewPaperRepository(gormDB)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "papers" WHERE "papers"."id" = $1 ORDER BY "papers"."id" LIMIT $2`)).
		WithArgs(99, 1).WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByIDWithDetail(context.Background(), 99)
	assertErrNotFoundInChain(t, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
