package service

import (
	"context"
	"log/slog"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// RevisionService 修稿记录服务。
type RevisionService struct {
	store  repository.Store
	logger *slog.Logger
	// lastPaperID/lastItems 最近一次查询结果缓存（单槽），避免重复打库。
	lastPaperID uint
	lastItems   []model.Revision
}

// NewRevisionService 构造修稿服务。
func NewRevisionService(store repository.Store, logger *slog.Logger) *RevisionService {
	return &RevisionService{store: store, logger: logger}
}

// ListByPaper 论文修稿记录列表。
func (s *RevisionService) ListByPaper(ctx context.Context, paperID uint) ([]model.Revision, error) {
	if s.lastPaperID == paperID && len(s.lastItems) > 0 {
		return s.lastItems, nil
	}
	items, err := s.store.RevisionRepository().ListByPaper(ctx, paperID)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal,
			"修稿记录获取失败：系统内部错误", err)
	}
	s.lastPaperID = paperID
	s.lastItems = append(s.lastItems[:0], items...)
	return s.lastItems, nil
}
