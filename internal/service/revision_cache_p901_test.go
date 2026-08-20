package service

import (
	"context"
	"sync"
	"testing"

	"github.com/paperflow/paperflow/internal/model"
)

// TestRevisionListConcurrentRaceP901 并发查询修稿记录，全程不得出现 data race。
func TestRevisionListConcurrentRaceP901(t *testing.T) {
	store := newFakeStore()
	svc := NewRevisionService(store, newTestLogger())
	store.revisions.revisions = []model.Revision{
		{PaperID: 1, Version: 1, FileName: "v1.pdf"},
		{PaperID: 2, Version: 1, FileName: "v2.pdf"},
	}
	ctx := context.Background()

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 50; j++ {
				if _, err := svc.ListByPaper(ctx, 1); err != nil {
					t.Errorf("list 1: %v", err)
					return
				}
				if _, err := svc.ListByPaper(ctx, 2); err != nil {
					t.Errorf("list 2: %v", err)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}

// TestRevisionListSnapshotStableP902 查询另一篇论文后，之前返回的修稿列表快照不得被改写。
func TestRevisionListSnapshotStableP902(t *testing.T) {
	store := newFakeStore()
	svc := NewRevisionService(store, newTestLogger())
	store.revisions.revisions = []model.Revision{
		{PaperID: 1, Version: 1, FileName: "v1.pdf"},
		{PaperID: 2, Version: 1, FileName: "v2.pdf"},
	}
	ctx := context.Background()

	first, err := svc.ListByPaper(ctx, 1)
	if err != nil {
		t.Fatalf("list 1: %v", err)
	}
	second, err := svc.ListByPaper(ctx, 2)
	if err != nil {
		t.Fatalf("list 2: %v", err)
	}
	if len(second) != 1 || second[0].FileName != "v2.pdf" {
		t.Fatalf("second = %+v, want v2.pdf", second)
	}
	if len(first) != 1 || first[0].FileName != "v1.pdf" {
		t.Fatalf("first snapshot polluted: %+v", first)
	}
}

// TestRevisionListHitIndependentP904 同一论文连续查询两次，两次返回的列表必须相互独立。
func TestRevisionListHitIndependentP904(t *testing.T) {
	store := newFakeStore()
	svc := NewRevisionService(store, newTestLogger())
	store.revisions.revisions = []model.Revision{{PaperID: 1, Version: 1, FileName: "v1.pdf"}}
	ctx := context.Background()

	first, err := svc.ListByPaper(ctx, 1)
	if err != nil {
		t.Fatalf("list 1: %v", err)
	}
	second, err := svc.ListByPaper(ctx, 1)
	if err != nil {
		t.Fatalf("list 1 again: %v", err)
	}
	second[0].FileName = "mutated"
	if first[0].FileName != "v1.pdf" {
		t.Fatalf("first list affected by second mutation: %s", first[0].FileName)
	}
}
