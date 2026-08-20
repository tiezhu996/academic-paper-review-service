package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"gorm.io/gorm"
)

// TestPlagiarismRepoChainErrNotFoundP803 FindByPaper 的 ErrNotFound 必须能被 errors.Is 识别。
func TestPlagiarismRepoChainErrNotFoundP803(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewPlagiarismRepository(gormDB)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "plagiarism_checks" WHERE paper_id = $1 ORDER BY "plagiarism_checks"."id" LIMIT $2`)).
		WithArgs(99, 1).WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByPaper(context.Background(), 99)
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
