package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"sync"
	"time"

	"github.com/paperflow/paperflow/internal/config"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// PlagiarismService 查重检测服务。
type PlagiarismService struct {
	store  repository.Store
	cfg    *config.Config
	logger *slog.Logger
	// recent 最近一次查重结果缓存：RunCheck 后写入，GetByPaper/Rerun 优先命中，避免重复打库与重复计算。
	// mu 保护 recent：编辑重跑查重（写）与读取查重结果（读）并发进行，裸 map 会触发
	// "concurrent map read and map write" 直接 panic 导致整个服务崩掉。
	recent map[uint]*model.PlagiarismCheck
	mu     sync.RWMutex
}

// NewPlagiarismService 构造查重服务。
func NewPlagiarismService(store repository.Store, cfg *config.Config, logger *slog.Logger) *PlagiarismService {
	return &PlagiarismService{store: store, cfg: cfg, logger: logger, recent: map[uint]*model.PlagiarismCheck{}}
}

// RunCheck 执行查重检测（确定性模拟），超过阈值自动退回论文。
// 该 service 方法同时被「论文创建自动查重」与「编辑手动重跑查重」两个接口复用。
func (s *PlagiarismService) RunCheck(ctx context.Context, paper *model.Paper) (*model.PlagiarismCheck, error) {
	sim := computeSimilarity(paper.Title + "|" + paper.Keywords)
	now := time.Now()
	check, err := s.store.PlagiarismRepository().FindByPaper(ctx, paper.ID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.ErrInternal, "查重检测失败：查询已有查重记录时系统内部错误", err)
		}
		check = &model.PlagiarismCheck{PaperID: paper.ID}
	}
	// 拷贝一份再写：recent 缓存与 handler 缓存都可能持有旧指针，
	// 若直接在原对象上改字段，重跑查重会把之前返回的快照原地改写，前端看到乱数据。
	check = cloneCheck(check)
	check.Similarity = sim
	check.Status = constants.PlagiarismStatusCompleted
	check.CheckedAt = &now
	check.Report = buildPlagiarismReport(sim, paper.Keywords)
	paper.Similarity = sim
	if sim > s.cfg.SimilarityThreshold {
		paper.Status = constants.PaperStatusRejected
		paper.FinalDecision = constants.ReviewDecisionReject
		paper.FinalComment = fmt.Sprintf("查重检测未通过：相似度 %s 超过阈值 %s，系统自动退回",
			util.FormatPercent(sim), util.FormatPercent(s.cfg.SimilarityThreshold))
		s.logger.Warn(fmt.Sprintf(constants.LogPlagiarismAutoReject, paper.ID, sim))
	}
	s.mu.Lock()
	s.recent[paper.ID] = check
	s.mu.Unlock()
	err = s.store.Transaction(ctx, func(tx repository.Store) error {
		if err := tx.PlagiarismRepository().Update(ctx, check); err != nil {
			return err
		}
		return tx.PaperRepository().Update(ctx, paper)
	})
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "查重检测失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogPlagiarismRun, paper.ID, sim))
	// 返回拷贝，避免调用方（如 handler 缓存）与 recent 持有同一指针而被后续重跑改写。
	return cloneCheck(check), nil
}

// GetByPaper 获取论文查重结果。
func (s *PlagiarismService) GetByPaper(ctx context.Context, paperID uint) (*model.PlagiarismCheck, error) {
	s.mu.RLock()
	c, ok := s.recent[paperID]
	s.mu.RUnlock()
	if ok {
		return cloneCheck(c), nil
	}
	check, err := s.store.PlagiarismRepository().FindByPaper(ctx, paperID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.ErrPlagiarismNotFound,
				fmt.Sprintf("查重结果获取失败：论文 id=%d 暂无查重记录", paperID), nil)
		}
		return nil, util.NewAppError(constants.ErrInternal, "查重结果获取失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogPlagiarismGet, paperID))
	return cloneCheck(check), nil
}

// cloneCheck 返回查重记录的深拷贝，隔离调用方持有的快照与缓存中的可变对象。
func cloneCheck(c *model.PlagiarismCheck) *model.PlagiarismCheck {
	if c == nil {
		return nil
	}
	cp := *c
	if c.CheckedAt != nil {
		t := *c.CheckedAt
		cp.CheckedAt = &t
	}
	return &cp
}

// Rerun 重新执行查重。
func (s *PlagiarismService) Rerun(ctx context.Context, paperID uint) (*model.PlagiarismCheck, error) {
	paper, err := s.store.PaperRepository().FindByID(ctx, paperID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.ErrPaperNotFound,
				fmt.Sprintf("查重重跑失败：论文 id=%d 不存在", paperID), nil)
		}
		return nil, util.NewAppError(constants.ErrInternal, "查重重跑失败：系统内部错误", err)
	}
	return s.RunCheck(ctx, paper)
}

// computeSimilarity 确定性模拟相似度：同一论文始终得到相同结果。
func computeSimilarity(text string) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(text))
	return 8 + float64(h.Sum32()%1800)/100.0 // 8.00 ~ 25.99
}

// buildPlagiarismReport 生成重复段落标注 JSON。
func buildPlagiarismReport(sim float64, keywords string) string {
	type reportItem struct {
		Source     string  `json:"source"`
		Paragraph  string  `json:"paragraph"`
		Similarity float64 `json:"similarity"`
	}
	kws := util.SplitKeywords(keywords)
	if len(kws) > 3 {
		kws = kws[:3]
	}
	items := make([]reportItem, 0, len(kws))
	for _, kw := range kws {
		items = append(items, reportItem{
			Source:     "公开文献库",
			Paragraph:  fmt.Sprintf("摘要中与「%s」相关的表述在公开文献中存在相似表达", kw),
			Similarity: sim / float64(len(kws)),
		})
	}
	b, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(b)
}
