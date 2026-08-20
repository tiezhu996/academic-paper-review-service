package service

import (
	"context"
	"sync"
	"testing"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/model"
)

// TestPlagiarismCacheConcurrentRaceP101 并发重跑查重 + 读取查重结果，全程不得出现 data race。
func TestPlagiarismCacheConcurrentRaceP101(t *testing.T) {
	store := newFakeStore()
	svc := NewPlagiarismService(store, newTestConfig(), newTestLogger())
	ctx := context.Background()

	paper := &model.Paper{Title: "race-paper-title", Keywords: "k1,k2", Status: constants.PaperStatusSubmitted}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	if err := store.plagiarism.Create(ctx, &model.PlagiarismCheck{PaperID: paper.ID, Status: constants.PlagiarismStatusPending}); err != nil {
		t.Fatalf("create check: %v", err)
	}
	if _, err := svc.RunCheck(ctx, paper); err != nil {
		t.Fatalf("initial run: %v", err)
	}

	const readers = 4
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for j := 0; j < 25; j++ {
			if _, err := svc.RunCheck(ctx, paper); err != nil {
				t.Errorf("concurrent runcheck: %v", err)
				return
			}
		}
	}()
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 60; j++ {
				got, err := svc.GetByPaper(ctx, paper.ID)
				if err != nil {
					t.Errorf("get by paper: %v", err)
					return
				}
				if got == nil || got.PaperID != paper.ID {
					t.Errorf("unexpected check: %+v", got)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}

// TestPlagiarismCacheSnapshotStableP102 重跑查重后，之前返回的查重结果快照不得被原地改写。
func TestPlagiarismCacheSnapshotStableP102(t *testing.T) {
	store := newFakeStore()
	svc := NewPlagiarismService(store, newTestConfig(), newTestLogger())
	ctx := context.Background()

	paper := &model.Paper{Title: "snapshot-paper-one", Keywords: "alpha,beta", Status: constants.PaperStatusSubmitted}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	if err := store.plagiarism.Create(ctx, &model.PlagiarismCheck{PaperID: paper.ID, Status: constants.PlagiarismStatusPending}); err != nil {
		t.Fatalf("create check: %v", err)
	}
	if _, err := svc.RunCheck(ctx, paper); err != nil {
		t.Fatalf("initial run: %v", err)
	}
	first, err := svc.GetByPaper(ctx, paper.ID)
	if err != nil {
		t.Fatalf("get by paper: %v", err)
	}
	sim1 := first.Similarity
	if sim1 <= 0 {
		t.Fatalf("expected positive similarity, got %v", sim1)
	}

	// 作者修改标题后重跑查重，相似度应变化；之前拿到的快照必须保持原值。
	paper.Title = "snapshot-paper-two-completely-different"
	if _, err := svc.RunCheck(ctx, paper); err != nil {
		t.Fatalf("rerun after title change: %v", err)
	}
	sim2 := computeSimilarity(paper.Title + "|" + paper.Keywords)
	if sim2 == sim1 {
		t.Fatalf("control failed: titles should produce different similarity (got %v)", sim1)
	}
	if first.Similarity != sim1 {
		t.Fatalf("snapshot mutated by later rerun: got %v, want %v", first.Similarity, sim1)
	}
}

