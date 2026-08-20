package handler

import (
	"context"
	"encoding/json"
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

type fakeUserRepoForList struct {
	repository.UserRepository
	users []model.User
}

func (f *fakeUserRepoForList) ListByRole(ctx context.Context, role string) ([]model.User, error) {
	return f.users, nil
}

type fakeAuditRepoForList struct {
	repository.AuditLogRepository
	logs []model.AuditLog
}

func (f *fakeAuditRepoForList) List(ctx context.Context, page, size int) ([]model.AuditLog, int64, error) {
	return f.logs, int64(len(f.logs)), nil
}

type fakeStoreForLists struct {
	repository.Store
	users *fakeUserRepoForList
	audit *fakeAuditRepoForList
}

func (f *fakeStoreForLists) UserRepository() repository.UserRepository   { return f.users }
func (f *fakeStoreForLists) AuditLogRepository() repository.AuditLogRepository { return f.audit }

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestUserListReviewersFreshP1002 审稿人数据变化后再次请求必须返回最新数据，不能返回过期缓存。
func TestUserListReviewersFreshP1002(t *testing.T) {
	store := &fakeStoreForLists{users: &fakeUserRepoForList{}, audit: &fakeAuditRepoForList{}}
	userSvc := service.NewUserService(store, newTestLogger())
	h := NewUserHandler(userSvc, newTestLogger())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users/reviewers", h.ListReviewers)

	store.users.users = []model.User{
		{ID: 1, Username: "r1", RealName: "审稿人一", Role: "reviewer"},
		{ID: 2, Username: "r2", RealName: "审稿人二", Role: "reviewer"},
	}
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/users/reviewers", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("first status = %d", w1.Code)
	}

	store.users.users = []model.User{{ID: 3, Username: "r3", RealName: "审稿人三", Role: "reviewer"}}
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/users/reviewers", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("second status = %d", w2.Code)
	}
	var resp struct {
		Data []*dtoUserSummary `json:"data"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, w2.Body.String())
	}
	if len(resp.Data) != 1 || resp.Data[0].Username != "r3" {
		t.Fatalf("second request returned stale reviewers: %+v", resp.Data)
	}
}

// TestAuditListNoInplaceMutationP1003 审计日志列表返回的 IP 必须保持原值，不能把共享数据就地改掉。
func TestAuditListNoInplaceMutationP1003(t *testing.T) {
	store := &fakeStoreForLists{users: &fakeUserRepoForList{}, audit: &fakeAuditRepoForList{}}
	auditSvc := service.NewAuditLogService(store, newTestLogger())
	h := NewAuditHandler(auditSvc, newTestLogger())
	store.audit.logs = []model.AuditLog{
		{ID: 1, Action: "login", Entity: "paper", EntityID: "1", IP: "1.2.3.4"},
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/audit-logs", h.List)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/audit-logs", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var resp struct {
		Data struct {
			Items []struct {
				IP string `json:"ip"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, w.Body.String())
	}
	if len(resp.Data.Items) != 1 || resp.Data.Items[0].IP != "1.2.3.4" {
		t.Fatalf("ip mutated in place: %+v", resp.Data.Items)
	}
}

type dtoUserSummary struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
}
