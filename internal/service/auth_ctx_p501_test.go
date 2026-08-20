package service

import (
	"context"
	"testing"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/util"

	"github.com/paperflow/paperflow/internal/dto"
)

// TestAuthRegisterCancelNoWriteP501 请求取消后注册必须中止，不能再继续创建用户。
func TestAuthRegisterCancelNoWriteP501(t *testing.T) {
	store := newFakeStore()
	svc := NewAuthService(store, newTestConfig(), newTestLogger())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Register(ctx, dto.RegisterRequest{
		Username: "alice", Password: "pass123", RealName: "Alice",
	})
	if err == nil {
		t.Fatalf("expected error when ctx already cancelled")
	}
	if _, ok := store.users.byName["alice"]; ok {
		t.Fatalf("user created despite cancelled ctx")
	}
}

// TestAuthLoginCancelNoWriteP505 请求取消后登录必须中止。
func TestAuthLoginCancelNoWriteP505(t *testing.T) {
	store := newFakeStore()
	svc := NewAuthService(store, newTestConfig(), newTestLogger())
	ctx := context.Background()
	hashed, err := util.HashPassword("pass123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := store.users.Create(ctx, &model.User{Username: "bob", Password: hashed, Role: constants.RoleAuthor}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err = svc.Login(ctx, dto.LoginRequest{Username: "bob", Password: "pass123"})
	if err == nil {
		t.Fatalf("expected error when ctx already cancelled")
	}
}
