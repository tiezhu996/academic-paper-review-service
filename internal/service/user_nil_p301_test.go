package service

import (
	"context"
	"testing"

	"github.com/paperflow/paperflow/internal/dto"
)

// TestUserMeMissingNoPanicP301 查询不存在的用户必须返回错误，不能返回 nil,nil。
func TestUserMeMissingNoPanicP301(t *testing.T) {
	store := newFakeStore()
	svc := NewUserService(store, newTestLogger())
	ctx := context.Background()

	user, err := svc.Me(ctx, 999)
	if err == nil {
		t.Fatalf("expected error for missing user, got user=%+v", user)
	}
}

// TestUserUpdateProfileMissingNoPanicP302 更新不存在的用户资料必须返回错误，不能 panic。
func TestUserUpdateProfileMissingNoPanicP302(t *testing.T) {
	store := newFakeStore()
	svc := NewUserService(store, newTestLogger())
	ctx := context.Background()

	_, err := svc.UpdateProfile(ctx, 999, dto.UpdateProfileRequest{Email: "someone@example.com"})
	if err == nil {
		t.Fatalf("expected error for missing user update")
	}
}
