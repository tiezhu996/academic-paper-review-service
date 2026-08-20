package service

import (
	"context"
	"testing"

	"github.com/paperflow/paperflow/internal/model"
)

// TestStatsTrendSnapshotStableP401 连续两次查询趋势，第一次结果不得被第二次改写。
func TestStatsTrendSnapshotStableP401(t *testing.T) {
	store := newFakeStore()
	svc := NewStatisticsService(store, newTestLogger())
	ctx := context.Background()

	store.papers.dayRows = []model.DayCount{{Day: "2026-08-01", Count: 3}}
	first, err := svc.Trend(ctx, 7)
	if err != nil {
		t.Fatalf("trend: %v", err)
	}

	store.papers.dayRows = []model.DayCount{{Day: "2026-08-02", Count: 9}}
	second, err := svc.Trend(ctx, 7)
	if err != nil {
		t.Fatalf("trend second: %v", err)
	}
	if len(second) != 1 || second[0].Count != 9 {
		t.Fatalf("second = %+v, want count 9", second)
	}
	if len(first) != 1 || first[0].Count != 3 || first[0].Day != "2026-08-01" {
		t.Fatalf("first snapshot polluted: %+v", first)
	}
}

// TestStatsSubjectsSnapshotStableP402 连续两次查询学科分布，第一次结果不得被第二次改写。
func TestStatsSubjectsSnapshotStableP402(t *testing.T) {
	store := newFakeStore()
	svc := NewStatisticsService(store, newTestLogger())
	ctx := context.Background()

	store.papers.subjectRows = []model.SubjectCount{{Subject: "computer", Count: 5}}
	first, err := svc.Subjects(ctx)
	if err != nil {
		t.Fatalf("subjects: %v", err)
	}

	store.papers.subjectRows = []model.SubjectCount{{Subject: "physics", Count: 7}}
	second, err := svc.Subjects(ctx)
	if err != nil {
		t.Fatalf("subjects second: %v", err)
	}
	if len(second) != 1 || second[0].Subject != "physics" {
		t.Fatalf("second = %+v, want physics", second)
	}
	if len(first) != 1 || first[0].Subject != "computer" || first[0].Count != 5 {
		t.Fatalf("first snapshot polluted: %+v", first)
	}
}

// TestStatsReviewerLoadSnapshotStableP403 连续两次查询审稿人工作量，第一次结果不得被第二次改写。
func TestStatsReviewerLoadSnapshotStableP403(t *testing.T) {
	store := newFakeStore()
	svc := NewStatisticsService(store, newTestLogger())
	ctx := context.Background()

	store.reviews.loadRows = []model.ReviewerLoad{{ReviewerID: 1, ReviewerName: "甲", Total: 4, Completed: 2}}
	first, err := svc.ReviewerWorkload(ctx)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	store.reviews.loadRows = []model.ReviewerLoad{{ReviewerID: 2, ReviewerName: "乙", Total: 8, Completed: 1}}
	second, err := svc.ReviewerWorkload(ctx)
	if err != nil {
		t.Fatalf("load second: %v", err)
	}
	if len(second) != 1 || second[0].ReviewerID != 2 {
		t.Fatalf("second = %+v, want reviewer 2", second)
	}
	if len(first) != 1 || first[0].ReviewerID != 1 || first[0].ReviewerName != "甲" {
		t.Fatalf("first snapshot polluted: %+v", first)
	}
}
