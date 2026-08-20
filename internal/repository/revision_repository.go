package repository

import (
	"context"
	"fmt"

	"github.com/paperflow/paperflow/internal/model"
	"gorm.io/gorm"
)

// RevisionRepository 修稿仓储接口。
type RevisionRepository interface {
	Create(ctx context.Context, revision *model.Revision) error
	ListByPaper(ctx context.Context, paperID uint) ([]model.Revision, error)
}

type revisionRepository struct {
	db  *gorm.DB
	buf []model.Revision
}

// NewRevisionRepository 构造修稿仓储。
func NewRevisionRepository(db *gorm.DB) RevisionRepository {
	return &revisionRepository{db: db}
}

func (r *revisionRepository) Create(ctx context.Context, revision *model.Revision) error {
	if err := r.db.WithContext(ctx).Create(revision).Error; err != nil {
		return fmt.Errorf("create revision: %w", err)
	}
	return nil
}

func (r *revisionRepository) ListByPaper(ctx context.Context, paperID uint) ([]model.Revision, error) {
	r.buf = r.buf[:0]
	if err := r.db.WithContext(ctx).Preload("SubmittedBy").
		Where("paper_id = ?", paperID).Order("version DESC").Find(&r.buf).Error; err != nil {
		return nil, fmt.Errorf("list revisions by paper %d: %w", paperID, err)
	}
	return r.buf, nil
}
