package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/chrono/pkg/cron"
)

func TestLoadJobsFile(t *testing.T) {
	yamlContent := `
jobs:
  - id: daily-report
    cron: "0 9 * * 1-5"
    target_url: http://api/report
    on_success: notify-slack
    on_failure: alert-admin
    max_retries: 3
    retry_delay_ms: 500
    retry_backoff_mult: 2.0
  - id: notify-slack
    target_url: http://slack/webhook
`

	dir := t.TempDir()
	filePath := filepath.Join(dir, "jobs.yaml")
	if err := os.WriteFile(filePath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp yaml file: %v", err)
	}

	jobs, err := LoadJobsFile(filePath)
	if err != nil {
		t.Fatalf("LoadJobsFile failed: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	j1 := jobs[0]
	if j1.ID != "daily-report" {
		t.Errorf("expected job 1 ID 'daily-report', got '%s'", j1.ID)
	}
	if j1.Cron != "0 9 * * 1-5" {
		t.Errorf("expected job 1 Cron '0 9 * * 1-5', got '%s'", j1.Cron)
	}
	if j1.TargetURL != "http://api/report" {
		t.Errorf("expected job 1 TargetURL 'http://api/report', got '%s'", j1.TargetURL)
	}
	if j1.OnSuccess != "notify-slack" {
		t.Errorf("expected job 1 OnSuccess 'notify-slack', got '%s'", j1.OnSuccess)
	}
	if j1.OnFailure != "alert-admin" {
		t.Errorf("expected job 1 OnFailure 'alert-admin', got '%s'", j1.OnFailure)
	}
	if j1.MaxRetries != 3 {
		t.Errorf("expected job 1 MaxRetries 3, got %d", j1.MaxRetries)
	}
	if j1.RetryDelayMs != 500 {
		t.Errorf("expected job 1 RetryDelayMs 500, got %d", j1.RetryDelayMs)
	}
	if j1.RetryBackoffMult != 2.0 {
		t.Errorf("expected job 1 RetryBackoffMult 2.0, got %f", j1.RetryBackoffMult)
	}

	j2 := jobs[1]
	if j2.ID != "notify-slack" {
		t.Errorf("expected job 2 ID 'notify-slack', got '%s'", j2.ID)
	}
	if j2.TargetURL != "http://slack/webhook" {
		t.Errorf("expected job 2 TargetURL 'http://slack/webhook', got '%s'", j2.TargetURL)
	}
}

func TestWatchJobsFile(t *testing.T) {
	// 1. Verify non-existent file returns error immediately
	err := WatchJobsFile("/path/to/nonexistent/jobs.yaml", func(j []cron.Job) {})
	if err == nil {
		t.Errorf("expected error when watching non-existent file, got nil")
	}

	// 2. Verify watching valid file detects updates
	dir := t.TempDir()
	filePath := filepath.Join(dir, "watched_jobs.yaml")
	initialContent := `
jobs:
  - id: initial-job
    target_url: http://initial/url
`
	if err := os.WriteFile(filePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	var mu sync.Mutex
	var updatedJobs []cron.Job
	stopChan := make(chan struct{})
	defer close(stopChan)

	err = WatchJobsFileWithInterval(filePath, 50*time.Millisecond, func(jobs []cron.Job) {
		mu.Lock()
		updatedJobs = jobs
		mu.Unlock()
	}, stopChan)

	if err != nil {
		t.Fatalf("WatchJobsFileWithInterval failed: %v", err)
	}

	// Ensure modtime changes by sleeping slightly before overwrite
	time.Sleep(100 * time.Millisecond)

	updatedContent := `
jobs:
  - id: updated-job
    target_url: http://updated/url
`
	if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("failed to write updated file: %v", err)
	}

	// Wait for watcher polling to trigger onChange
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		count := len(updatedJobs)
		mu.Unlock()
		if count > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(updatedJobs) != 1 {
		t.Fatalf("expected 1 updated job, got %d", len(updatedJobs))
	}
	if updatedJobs[0].ID != "updated-job" {
		t.Errorf("expected updated job ID 'updated-job', got '%s'", updatedJobs[0].ID)
	}
}
