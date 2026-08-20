package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/paperflow/paperflow/internal/model"
)

// TestUserRepoCreateCtxCancelP503 上下文取消后创建用户必须中止，不能绕开取消继续写库。
func TestUserRepoCreateCtxCancelP503(t *testing.T) {
	gormDB, _ := newMockDB(t)
	repo := NewUserRepository(gormDB)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Create(ctx, &model.User{Username: "alice", Password: "x", Role: "author"})
	if err == nil {
		t.Fatalf("expected error on cancelled ctx")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// TestUserRepoFindByUsernameCtxCancelP504 上下文取消后查询用户必须中止。
func TestUserRepoFindByUsernameCtxCancelP504(t *testing.T) {
	gormDB, _ := newMockDB(t)
	repo := NewUserRepository(gormDB)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.FindByUsername(ctx, "alice")
	if err == nil {
		t.Fatalf("expected error on cancelled ctx")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// TestUserRepoFindByIDCtxCancelP506 上下文取消后按 ID 查询用户必须中止。
func TestUserRepoFindByIDCtxCancelP506(t *testing.T) {
	gormDB, _ := newMockDB(t)
	repo := NewUserRepository(gormDB)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.FindByID(ctx, 1)
	if err == nil {
		t.Fatalf("expected error on cancelled ctx")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
