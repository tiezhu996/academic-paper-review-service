package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paperflow/paperflow/internal/config"
)

// TestStorageUploadErrorSurfacesP602 对象存储上传失败必须透出错误，不能吞成成功。
func TestStorageUploadErrorSurfacesP602(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &config.Config{
		MinIOEndpoint:  strings.TrimPrefix(srv.URL, "http://"),
		MinIOAccessKey: "minioadmin",
		MinIOSecretKey: "minioadmin",
		MinIOBucket:    "paperflow-files",
		MinIOUseSSL:    false,
	}
	svc, err := NewStorageService(cfg, newTestLogger())
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	err = svc.Upload(context.Background(), "papers/x.pdf", strings.NewReader("data"), 4, "application/pdf")
	if err == nil {
		t.Fatalf("expected error when object store upload fails")
	}
}

// TestStorageEnsureBucketErrorSurfacesP603 创建桶失败必须透出错误，不能吞成成功。
func TestStorageEnsureBucketErrorSurfacesP603(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &config.Config{
		MinIOEndpoint:  strings.TrimPrefix(srv.URL, "http://"),
		MinIOAccessKey: "minioadmin",
		MinIOSecretKey: "minioadmin",
		MinIOBucket:    "paperflow-files",
		MinIOUseSSL:    false,
	}
	svc, err := NewStorageService(cfg, newTestLogger())
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	err = svc.EnsureBucket(context.Background())
	if err == nil {
		t.Fatalf("expected error when bucket create fails")
	}
}
