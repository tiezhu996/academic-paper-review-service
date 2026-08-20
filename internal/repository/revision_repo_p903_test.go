package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestRevisionRepoListBufferNoPolluteP903 修稿记录仓储连续查询两次，第一次结果不得被第二次改写。
func TestRevisionRepoListBufferNoPolluteP903(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewRevisionRepository(gormDB)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM "revisions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "paper_id", "version", "file_name", "submitted_by_id"}).
			AddRow(1, 1, 1, "v1.pdf", 7))
	mock.ExpectQuery(`SELECT .* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(7, "author"))

	first, err := repo.ListByPaper(ctx, 1)
	if err != nil {
		t.Fatalf("first list: %v", err)
	}

	mock.ExpectQuery(`SELECT .* FROM "revisions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "paper_id", "version", "file_name", "submitted_by_id"}).
			AddRow(2, 1, 2, "v2.pdf", 7))
	mock.ExpectQuery(`SELECT .* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(7, "author"))

	second, err := repo.ListByPaper(ctx, 1)
	if err != nil {
		t.Fatalf("second list: %v", err)
	}
	if len(second) != 1 || second[0].FileName != "v2.pdf" {
		t.Fatalf("second = %+v, want v2.pdf", second)
	}
	if len(first) != 1 || first[0].FileName != "v1.pdf" {
		t.Fatalf("first list polluted: %+v", first)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
