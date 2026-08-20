package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/service"
	"github.com/paperflow/paperflow/internal/util"
)

// PlagiarismHandler 查重处理器。
type PlagiarismHandler struct {
	plagiarismSvc *service.PlagiarismService
	auditSvc      *service.AuditLogService
	logger        *slog.Logger
	// recent 最近返回给前端的查重结果缓存：GetByPaper 命中直接返回，避免重复调用服务。
	recent map[uint]*model.PlagiarismCheck
}

// NewPlagiarismHandler 构造查重处理器。
func NewPlagiarismHandler(plagiarismSvc *service.PlagiarismService, auditSvc *service.AuditLogService, logger *slog.Logger) *PlagiarismHandler {
	return &PlagiarismHandler{plagiarismSvc: plagiarismSvc, auditSvc: auditSvc, logger: logger, recent: map[uint]*model.PlagiarismCheck{}}
}

// GetByPaper 获取论文查重结果。
func (h *PlagiarismHandler) GetByPaper(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if cached, ok := h.recent[id]; ok {
		util.OK(c, cached)
		return
	}
	check, err := h.plagiarismSvc.GetByPaper(c.Request.Context(), id)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	h.recent[id] = check
	util.OK(c, check)
}

// Rerun 重跑查重（编辑/管理员）。
func (h *PlagiarismHandler) Rerun(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	check, err := h.plagiarismSvc.Rerun(c.Request.Context(), id)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	if e := h.auditSvc.Record(c.Request.Context(), util.GetUserID(c), util.GetUsername(c),
		constants.AuditActionRunPlagiarism, "paper", fmt.Sprint(id), "重跑查重", c.ClientIP(), util.GetRequestID(c)); e != nil {
		h.logger.Error("audit rerun plagiarism failed", "error", e)
	}
	util.OK(c, check)
}

// wrapError 查重处理器错误包装。
func (h *PlagiarismHandler) wrapError(c *gin.Context, err error) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		util.Fail(c, httpStatusForCode(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "系统内部错误："+err.Error())
}
