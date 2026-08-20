package handler

import (
	"context"
	"io"
	"log/slog"

	"github.com/paperflow/paperflow/internal/config"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
)

// 供 handler 层测试使用的最小内存仓储实现（只实现本组测试用到的接口）。

type fakePlagRepo struct {
	repository.PlagiarismRepository
	checks map[uint]*model.PlagiarismCheck
	nextID uint
}

func (f *fakePlagRepo) Create(ctx context.Context, c *model.PlagiarismCheck) error {
	f.nextID++
	c.ID = f.nextID
	f.checks[c.ID] = c
	return nil
}

func (f *fakePlagRepo) Update(ctx context.Context, c *model.PlagiarismCheck) error {
	f.checks[c.ID] = c
	return nil
}

func (f *fakePlagRepo) FindByPaper(ctx context.Context, paperID uint) (*model.PlagiarismCheck, error) {
	for _, c := range f.checks {
		if c.PaperID == paperID {
			return c, nil
		}
	}
	return nil, repository.ErrNotFound
}

type fakePaperRepo struct {
	repository.PaperRepository
	papers map[uint]*model.Paper
	nextID uint
}

func (f *fakePaperRepo) Create(ctx context.Context, p *model.Paper) error {
	f.nextID++
	p.ID = f.nextID
	f.papers[p.ID] = p
	return nil
}

func (f *fakePaperRepo) Update(ctx context.Context, p *model.Paper) error {
	f.papers[p.ID] = p
	return nil
}

func (f *fakePaperRepo) FindByID(ctx context.Context, id uint) (*model.Paper, error) {
	if p, ok := f.papers[id]; ok {
		return p, nil
	}
	return nil, repository.ErrNotFound
}

type fakeAuditRepo struct {
	repository.AuditLogRepository
	logs []model.AuditLog
}

func (f *fakeAuditRepo) Create(ctx context.Context, l *model.AuditLog) error {
	f.logs = append(f.logs, *l)
	return nil
}

type fakeStore struct {
	repository.Store
	plag   *fakePlagRepo
	papers *fakePaperRepo
	audit  *fakeAuditRepo
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		plag:   &fakePlagRepo{checks: map[uint]*model.PlagiarismCheck{}},
		papers: &fakePaperRepo{papers: map[uint]*model.Paper{}},
		audit:  &fakeAuditRepo{},
	}
}

func (f *fakeStore) Transaction(ctx context.Context, fn func(repository.Store) error) error {
	return fn(f)
}

func (f *fakeStore) PlagiarismRepository() repository.PlagiarismRepository { return f.plag }
func (f *fakeStore) PaperRepository() repository.PaperRepository           { return f.papers }
func (f *fakeStore) AuditLogRepository() repository.AuditLogRepository     { return f.audit }

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestConfig() *config.Config {
	return &config.Config{JWTSecret: "test-secret", JWTExpireHours: 72, SimilarityThreshold: 30}
}
