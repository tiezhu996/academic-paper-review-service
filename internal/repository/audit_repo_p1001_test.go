package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestAuditRepoListBufferNoPolluteP1001 审计日志仓储连续查询两次，第一次结果不得被第二次改写。
func TestAuditRepoListBufferNoPolluteP1001(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewAuditLogRepository(gormDB)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "audit_logs"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT .* FROM "audit_logs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "action", "ip"}).
			AddRow(1, "login", "1.1.1.1").AddRow(2, "logout", "2.2.2.2"))

	first, _, err := repo.List(ctx, 1, 10)
	if err != nil {
		t.Fatalf("first list: %v", err)
	}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "audit_logs"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT .* FROM "audit_logs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "action", "ip"}).
			AddRow(3, "login", "3.3.3.3").AddRow(4, "logout", "4.4.4.4"))

	second, _, err := repo.List(ctx, 1, 10)
	if err != nil {
		t.Fatalf("second list: %v", err)
	}
	if len(second) != 2 || second[0].ID != 3 {
		t.Fatalf("second = %+v, want id 3", second)
	}
	if len(first) != 2 || first[0].ID != 1 {
		t.Fatalf("first list polluted: %+v", first)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestAuditRepoListTotalP1004 审计日志列表 total 必须返回真实总数，不能只返回当前页条数。
func TestAuditRepoListTotalP1004(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewAuditLogRepository(gormDB)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "audit_logs"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))
	rows := sqlmock.NewRows([]string{"id", "action", "ip"})
	for i := 1; i <= 10; i++ {
		rows = rows.AddRow(i, "login", "1.2.3.4")
	}
	mock.ExpectQuery(`SELECT .* FROM "audit_logs"`).WillReturnRows(rows)

	_, total, err := repo.List(ctx, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 25 {
		t.Fatalf("total = %d, want 25", total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
