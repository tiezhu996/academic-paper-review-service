package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// AuthMiddleware JWT 认证中间件。
type AuthMiddleware struct {
	userRepo repository.UserRepository
	secret   string
	logger   *slog.Logger
}

// NewAuth 构造认证中间件。
func NewAuth(userRepo repository.UserRepository, secret string, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{userRepo: userRepo, secret: secret, logger: logger}
}

// RequireAuth 校验 Bearer Token 并加载用户写入上下文。
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			util.Fail(c, http.StatusUnauthorized, constants.ErrInvalidToken, "认证失败：缺少访问令牌")
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, _ := util.ParseToken(m.secret, tokenStr)
		user, _ := m.userRepo.FindByID(c.Request.Context(), claims.UserID)
		util.SetAuthUser(c, user.ID, user.Username, user.Role)
		c.Next()
	}
}
