package cron

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestDAGChainLinear tests A -> B -> C execution on success.
func TestDAGChainLinear(t *testing.T) {
	var executed sync.Map

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Query().Get("id")
		executed.Store(jobID, time.Now())
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sched := NewScheduler(nil)

	jobC := &Job{
		ID:        "job-c",
		TargetURL: server.URL + "?id=job-c",
	}
	jobB := &Job{
		ID:        "job-b",
		TargetURL: server.URL + "?id=job-b",
		OnSuccess: "job-c",
	}
	jobA := &Job{
		ID:        "job-a",
		TargetURL: server.URL + "?id=job-a",
		OnSuccess: "job-b",
	}

	_ = sched.AddJob(jobC)
	_ = sched.AddJob(jobB)
	_ = sched.AddJob(jobA)

	err := sched.TriggerJob("job-a")
	if err != nil {
		t.Fatalf("failed to trigger job-a: %v", err)
	}

	// Wait for chain completion
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_, okA := executed.Load("job-a")
		_, okB := executed.Load("job-b")
		_, okC := executed.Load("job-c")
		if okA && okB && okC {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	_, okA := executed.Load("job-a")
	_, okB := executed.Load("job-b")
	_, okC := executed.Load("job-c")
	t.Errorf("linear chain failed: job-a=%v, job-b=%v, job-c=%v", okA, okB, okC)
}

// TestDAGChainFailureBranch tests A-fail -> B execution on failure.
func TestDAGChainFailureBranch(t *testing.T) {
	var executed sync.Map

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Query().Get("id")
		executed.Store(jobID, time.Now())
		if jobID == "job-a" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	sched := NewScheduler(nil)

	jobB := &Job{
		ID:        "job-b",
		TargetURL: server.URL + "?id=job-b",
	}
	jobA := &Job{
		ID:        "job-a",
		TargetURL: server.URL + "?id=job-a",
		OnFailure: "job-b",
	}

	_ = sched.AddJob(jobB)
	_ = sched.AddJob(jobA)

	err := sched.TriggerJob("job-a")
	if err != nil {
		t.Fatalf("failed to trigger job-a: %v", err)
	}

	// Wait for failure branch completion
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_, okA := executed.Load("job-a")
		_, okB := executed.Load("job-b")
		if okA && okB {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	_, okA := executed.Load("job-a")
	_, okB := executed.Load("job-b")
	if !okA || !okB {
		t.Fatalf("failure branch failed to execute: job-a=%v, job-b=%v", okA, okB)
	}

	sched.mu.RLock()
	jA := sched.jobs["job-a"]
	jB := sched.jobs["job-b"]
	sched.mu.RUnlock()

	if jA.LastOutcome != "failed" {
		t.Errorf("expected job-a outcome 'failed', got '%s'", jA.LastOutcome)
	}
	if jB.LastOutcome != "success" {
		t.Errorf("expected job-b outcome 'success', got '%s'", jB.LastOutcome)
	}
}

// TestDAGChainCycleGuard tests cycle guard A -> B -> A stopping at max depth 10.
func TestDAGChainCycleGuard(t *testing.T) {
	var count int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sched := NewScheduler(nil)

	jobB := &Job{
		ID:        "job-b",
		TargetURL: server.URL,
		OnSuccess: "job-a",
	}
	jobA := &Job{
		ID:        "job-a",
		TargetURL: server.URL,
		OnSuccess: "job-b",
	}

	_ = sched.AddJob(jobB)
	_ = sched.AddJob(jobA)

	_ = sched.TriggerJob("job-a")

	// Wait enough time for chain to complete up to depth 10
	time.Sleep(500 * time.Millisecond)

	totalExecutions := atomic.LoadInt32(&count)
	if totalExecutions != 10 {
		t.Errorf("expected cycle guard to stop at 10 executions, got %d", totalExecutions)
	}
}

// TestRetrySuccessfulOnSecondAttempt tests job retrying and succeeding on 2nd attempt.
func TestRetrySuccessfulOnSecondAttempt(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	sched := NewScheduler(nil)

	job := &Job{
		ID:               "retry-job",
		TargetURL:        server.URL,
		MaxRetries:       2,
		RetryDelayMs:     10,
		RetryBackoffMult: 1.0,
	}

	_ = sched.AddJob(job)
	_ = sched.TriggerJob("retry-job")

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&attempts) >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	sched.mu.RLock()
	j := sched.jobs["retry-job"]
	sched.mu.RUnlock()

	if j.LastOutcome != "success" {
		t.Errorf("expected LastOutcome 'success', got '%s'", j.LastOutcome)
	}
	if j.FailureCount != 0 {
		t.Errorf("expected FailureCount 0, got %d", j.FailureCount)
	}
	if j.RetryCount != 1 {
		t.Errorf("expected RetryCount 1 (1 retry attempt), got %d", j.RetryCount)
	}
}

// TestRetryExhaustedRetries tests exhausted retries incrementing FailureCount.
func TestRetryExhaustedRetries(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sched := NewScheduler(nil)

	job := &Job{
		ID:               "failing-job",
		TargetURL:        server.URL,
		MaxRetries:       2,
		RetryDelayMs:     10,
		RetryBackoffMult: 1.0,
	}

	_ = sched.AddJob(job)
	_ = sched.TriggerJob("failing-job")

	time.Sleep(300 * time.Millisecond)

	sched.mu.RLock()
	j := sched.jobs["failing-job"]
	sched.mu.RUnlock()

	total := atomic.LoadInt32(&attempts)
	if total != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 total HTTP attempts, got %d", total)
	}
	if j.LastOutcome != "failed" {
		t.Errorf("expected LastOutcome 'failed', got '%s'", j.LastOutcome)
	}
	if j.FailureCount != 1 {
		t.Errorf("expected FailureCount 1, got %d", j.FailureCount)
	}
	if j.RetryCount != 2 {
		t.Errorf("expected RetryCount 2, got %d", j.RetryCount)
	}
}

// TestRetryJitterNonDeterministicDelays tests jitter producing variation within ±10%.
func TestRetryJitterNonDeterministicDelays(t *testing.T) {
	baseDelay := 100
	backoff := 2.0
	attempt := 1 // base calculation: 100 * 2^1 = 200ms, jitter +-10% => [180ms, 220ms]

	var delays []time.Duration
	for i := 0; i < 50; i++ {
		d := CalculateRetryDelay(baseDelay, backoff, attempt)
		if d < 180*time.Millisecond || d > 220*time.Millisecond {
			t.Fatalf("delay %v out of jitter bounds [180ms, 220ms]", d)
		}
		delays = append(delays, d)
	}

	hasDifference := false
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			hasDifference = true
			break
		}
	}

	if !hasDifference {
		t.Errorf("expected non-deterministic jitter delays, but all 50 samples were identical: %v", delays[0])
	}
}
