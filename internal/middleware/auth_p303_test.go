package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

type fakeUserRepo struct {
	repository.UserRepository
	users map[uint]*model.User
}

func (f *fakeUserRepo) FindByID(ctx context.Context, id uint) (*model.User, error) {
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newAuthEngine(t *testing.T, repo *fakeUserRepo, secret string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mw := NewAuth(repo, secret, newTestLogger())
	e := gin.New()
	e.Use(gin.Recovery())
	e.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"uid": util.GetUserID(c)})
	})
	return e
}

// TestAuthInvalidTokenNoPanicP303 无效 token 必须回 401，不能因解引用 nil 崩溃。
func TestAuthInvalidTokenNoPanicP303(t *testing.T) {
	repo := &fakeUserRepo{users: map[uint]*model.User{}}
	e := newAuthEngine(t, repo, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

// TestAuthDeletedUserNoPanicP304 用户被删除后携带旧 token 必须回 401，不能因解引用 nil 崩溃。
func TestAuthDeletedUserNoPanicP304(t *testing.T) {
	repo := &fakeUserRepo{users: map[uint]*model.User{}}
	const secret = "test-secret"
	e := newAuthEngine(t, repo, secret)

	token, err := util.GenerateToken(secret, 72, 999, "ghost", "author")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
