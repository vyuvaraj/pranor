package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vyuvaraj/serv/packages/ServCron/pkg/config"
	"github.com/vyuvaraj/serv/packages/ServCron/pkg/cron"
)

// ============================================================================
// CR.G1: DAG Job Chain Pipeline Tests
// ============================================================================

// --- Tier 1: Feature Coverage (CR.G1) ---

func TestE2E_CR_G1_Tier1_LinearChain_A_B_C(t *testing.T) {
	var callOrder []string
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Query().Get("id")
		mu.Lock()
		callOrder = append(callOrder, jobID)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)

	jobC := &cron.Job{ID: "jobC", TargetURL: ts.URL + "?id=jobC", Status: "active"}
	jobB := &cron.Job{ID: "jobB", TargetURL: ts.URL + "?id=jobB", Status: "active", OnSuccess: "jobC"}
	jobA := &cron.Job{ID: "jobA", TargetURL: ts.URL + "?id=jobA", Status: "active", OnSuccess: "jobB"}

	_ = sched.AddJob(jobC)
	_ = sched.AddJob(jobB)
	_ = sched.AddJob(jobA)

	if err := sched.TriggerJob("jobA"); err != nil {
		t.Fatalf("TriggerJob failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(callOrder) != 3 {
		t.Fatalf("expected 3 job calls in chain, got %d (%v)", len(callOrder), callOrder)
	}
	if callOrder[0] != "jobA" || callOrder[1] != "jobB" || callOrder[2] != "jobC" {
		t.Errorf("expected chain order [jobA, jobB, jobC], got %v", callOrder)
	}
}

func TestE2E_CR_G1_Tier1_FailureBranch_A_Fail_B(t *testing.T) {
	var executed []string
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Query().Get("id")
		mu.Lock()
		executed = append(executed, jobID)
		mu.Unlock()

		if jobID == "jobA" {
			w.WriteHeader(http.StatusInternalServerError) // Job A fails
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)

	jobB := &cron.Job{ID: "jobB", TargetURL: ts.URL + "?id=jobB", Status: "active"}
	jobA := &cron.Job{ID: "jobA", TargetURL: ts.URL + "?id=jobA", Status: "active", OnSuccess: "jobSuccess", OnFailure: "jobB"}

	_ = sched.AddJob(jobB)
	_ = sched.AddJob(jobA)

	_ = sched.TriggerJob("jobA")
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(executed) != 2 || executed[0] != "jobA" || executed[1] != "jobB" {
		t.Errorf("expected [jobA, jobB] on failure branch, got %v", executed)
	}
}

func TestE2E_CR_G1_Tier1_SuccessBranch_IgnoreFailure(t *testing.T) {
	var executed []string
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Query().Get("id")
		mu.Lock()
		executed = append(executed, jobID)
		mu.Unlock()
		w.WriteHeader(http.StatusOK) // Job A succeeds
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)

	jobSuccess := &cron.Job{ID: "jobSuccess", TargetURL: ts.URL + "?id=jobSuccess", Status: "active"}
	jobFail := &cron.Job{ID: "jobFail", TargetURL: ts.URL + "?id=jobFail", Status: "active"}
	jobA := &cron.Job{
		ID:        "jobA",
		TargetURL: ts.URL + "?id=jobA",
		Status:    "active",
		OnSuccess: "jobSuccess",
		OnFailure: "jobFail",
	}

	_ = sched.AddJob(jobSuccess)
	_ = sched.AddJob(jobFail)
	_ = sched.AddJob(jobA)

	_ = sched.TriggerJob("jobA")
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(executed) != 2 || executed[0] != "jobA" || executed[1] != "jobSuccess" {
		t.Errorf("expected [jobA, jobSuccess], got %v", executed)
	}
}

func TestE2E_CR_G1_Tier1_CycleGuard_MaxDepth10(t *testing.T) {
	var callCount int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)

	// Self-referencing job loop: jobSelf -> jobSelf
	jobSelf := &cron.Job{
		ID:        "jobSelf",
		TargetURL: ts.URL,
		Status:    "active",
		OnSuccess: "jobSelf",
	}

	_ = sched.AddJob(jobSelf)
	_ = sched.TriggerJob("jobSelf")

	time.Sleep(500 * time.Millisecond)

	count := atomic.LoadInt32(&callCount)
	if count > 10 {
		t.Errorf("cycle guard failed: expected max 10 calls, got %d", count)
	}
	if count < 10 {
		t.Errorf("expected 10 executions up to max chain depth, got %d", count)
	}
}

func TestE2E_CR_G1_Tier1_InactiveTargetJob_NotTriggered(t *testing.T) {
	var executed []string
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Query().Get("id")
		mu.Lock()
		executed = append(executed, jobID)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)

	jobB := &cron.Job{ID: "jobB", TargetURL: ts.URL + "?id=jobB", Status: "paused"} // paused status
	jobA := &cron.Job{ID: "jobA", TargetURL: ts.URL + "?id=jobA", Status: "active", OnSuccess: "jobB"}

	_ = sched.AddJob(jobB)
	_ = sched.AddJob(jobA)

	_ = sched.TriggerJob("jobA")
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(executed) != 1 || executed[0] != "jobA" {
		t.Errorf("paused jobB should not be triggered, got executions: %v", executed)
	}
}

// --- Tier 2: Boundary & Corner Cases (CR.G1) ---

func TestE2E_CR_G1_Tier2_NonExistentTargetJob(t *testing.T) {
	var callCount int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)
	jobA := &cron.Job{ID: "jobA", TargetURL: ts.URL, Status: "active", OnSuccess: "ghost_job"}

	_ = sched.AddJob(jobA)
	_ = sched.TriggerJob("jobA")

	time.Sleep(150 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 execution for jobA, non-existent target should do nothing")
	}
}

func TestE2E_CR_G1_Tier2_EmptyOnSuccessOnFailure(t *testing.T) {
	var callCount int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)
	jobA := &cron.Job{ID: "jobA", TargetURL: ts.URL, Status: "active", OnSuccess: "", OnFailure: ""}

	_ = sched.AddJob(jobA)
	_ = sched.TriggerJob("jobA")

	time.Sleep(150 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 execution for empty chain fields")
	}
}

func TestE2E_CR_G1_Tier2_MutualRecursionCycle(t *testing.T) {
	var callCount int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)

	job1 := &cron.Job{ID: "job1", TargetURL: ts.URL, Status: "active", OnSuccess: "job2"}
	job2 := &cron.Job{ID: "job2", TargetURL: ts.URL, Status: "active", OnSuccess: "job1"}

	_ = sched.AddJob(job1)
	_ = sched.AddJob(job2)

	_ = sched.TriggerJob("job1")
	time.Sleep(500 * time.Millisecond)

	count := atomic.LoadInt32(&callCount)
	if count > 10 {
		t.Errorf("mutual recursion guard failed: expected max 10 calls, got %d", count)
	}
}

func TestE2E_CR_G1_Tier2_DiamondDAG(t *testing.T) {
	// Job A -> triggers B; Job B -> triggers C
	var executed []string
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		mu.Lock()
		executed = append(executed, id)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)

	jC := &cron.Job{ID: "C", TargetURL: ts.URL + "?id=C", Status: "active"}
	jB := &cron.Job{ID: "B", TargetURL: ts.URL + "?id=B", Status: "active", OnSuccess: "C"}
	jA := &cron.Job{ID: "A", TargetURL: ts.URL + "?id=A", Status: "active", OnSuccess: "B"}

	_ = sched.AddJob(jC)
	_ = sched.AddJob(jB)
	_ = sched.AddJob(jA)

	_ = sched.TriggerJob("A")
	time.Sleep(350 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(executed) != 3 || executed[0] != "A" || executed[1] != "B" || executed[2] != "C" {
		t.Errorf("expected [A, B, C], got %v", executed)
	}
}

func TestE2E_CR_G1_Tier2_ChainExecution_ThreadSafety(t *testing.T) {
	var totalExecs int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&totalExecs, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)

	for i := 0; i < 5; i++ {
		j2 := &cron.Job{ID: fmt.Sprintf("chain_%d_2", i), TargetURL: ts.URL, Status: "active"}
		j1 := &cron.Job{ID: fmt.Sprintf("chain_%d_1", i), TargetURL: ts.URL, Status: "active", OnSuccess: j2.ID}
		_ = sched.AddJob(j2)
		_ = sched.AddJob(j1)
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = sched.TriggerJob(fmt.Sprintf("chain_%d_1", idx))
		}(i)
	}
	wg.Wait()

	time.Sleep(400 * time.Millisecond)

	if atomic.LoadInt32(&totalExecs) != 10 {
		t.Errorf("expected 10 total executions across 5 parallel chains, got %d", atomic.LoadInt32(&totalExecs))
	}
}

// ============================================================================
// CR.G2: Per-Job Retry Policy Engine Tests
// ============================================================================

// --- Tier 1: Feature Coverage (CR.G2) ---

func TestE2E_CR_G2_Tier1_CalculateRetryDelay_Exponential(t *testing.T) {
	baseDelay := 100
	backoffMult := 2.0

	// Attempt 0: ~100ms
	d0 := cron.CalculateRetryDelay(baseDelay, backoffMult, 0)
	if d0 < 80*time.Millisecond || d0 > 120*time.Millisecond {
		t.Errorf("attempt 0 delay out of bounds: %v", d0)
	}

	// Attempt 1: ~200ms
	d1 := cron.CalculateRetryDelay(baseDelay, backoffMult, 1)
	if d1 < 160*time.Millisecond || d1 > 240*time.Millisecond {
		t.Errorf("attempt 1 delay out of bounds: %v", d1)
	}

	// Attempt 2: ~400ms
	d2 := cron.CalculateRetryDelay(baseDelay, backoffMult, 2)
	if d2 < 320*time.Millisecond || d2 > 480*time.Millisecond {
		t.Errorf("attempt 2 delay out of bounds: %v", d2)
	}
}

func TestE2E_CR_G2_Tier1_SuccessfulRetry_2ndAttempt(t *testing.T) {
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cnt := atomic.AddInt32(&attempts, 1)
		if cnt == 1 {
			w.WriteHeader(http.StatusInternalServerError) // fail 1st attempt
		} else {
			w.WriteHeader(http.StatusOK) // succeed 2nd attempt
		}
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)
	job := &cron.Job{
		ID:               "retry_job_2nd",
		TargetURL:        ts.URL,
		Status:           "active",
		MaxRetries:       3,
		RetryDelayMs:     10,
		RetryBackoffMult: 1.0,
	}

	_ = sched.AddJob(job)
	_ = sched.TriggerJob("retry_job_2nd")

	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected exactly 2 attempts (1 initial + 1 retry), got %d", atomic.LoadInt32(&attempts))
	}

	jobs := sched.GetJobs()
	for _, j := range jobs {
		if j.ID == "retry_job_2nd" {
			if j.LastOutcome != "success" {
				t.Errorf("expected LastOutcome success, got %q", j.LastOutcome)
			}
			if j.FailureCount != 0 {
				t.Errorf("expected FailureCount 0 on success, got %d", j.FailureCount)
			}
		}
	}
}

func TestE2E_CR_G2_Tier1_ExhaustedRetries_FailureCount(t *testing.T) {
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError) // always fail
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)
	job := &cron.Job{
		ID:               "retry_job_exhaust",
		TargetURL:        ts.URL,
		Status:           "active",
		MaxRetries:       2,
		RetryDelayMs:     5,
		RetryBackoffMult: 1.0,
	}

	_ = sched.AddJob(job)
	_ = sched.TriggerJob("retry_job_exhaust")

	time.Sleep(300 * time.Millisecond)

	// 1 initial + 2 retries = 3 attempts total
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 total attempts for MaxRetries=2, got %d", atomic.LoadInt32(&attempts))
	}

	jobs := sched.GetJobs()
	for _, j := range jobs {
		if j.ID == "retry_job_exhaust" {
			if j.LastOutcome != "failed" {
				t.Errorf("expected LastOutcome failed, got %q", j.LastOutcome)
			}
			if j.FailureCount != 1 {
				t.Errorf("expected FailureCount 1 after exhaustion, got %d", j.FailureCount)
			}
		}
	}
}

func TestE2E_CR_G2_Tier1_Jitter_NonDeterministic(t *testing.T) {
	// Verify CalculateRetryDelay generates varying delays due to jitter
	baseDelay := 100
	backoff := 1.5

	delays := make(map[time.Duration]bool)
	for i := 0; i < 50; i++ {
		d := cron.CalculateRetryDelay(baseDelay, backoff, 2)
		delays[d] = true
	}

	if len(delays) < 10 {
		t.Errorf("expected non-deterministic delays from jitter, got only %d distinct values", len(delays))
	}
}

func TestE2E_CR_G2_Tier1_ZeroRetries_ImmediateFail(t *testing.T) {
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)
	job := &cron.Job{
		ID:         "no_retry_job",
		TargetURL:  ts.URL,
		Status:     "active",
		MaxRetries: 0,
	}

	_ = sched.AddJob(job)
	_ = sched.TriggerJob("no_retry_job")

	time.Sleep(150 * time.Millisecond)

	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected 1 attempt for MaxRetries=0, got %d", atomic.LoadInt32(&attempts))
	}
}

// --- Tier 2: Boundary & Corner Cases (CR.G2) ---

func TestE2E_CR_G2_Tier2_ZeroDelay_FastRetry(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)
	job := &cron.Job{
		ID:           "zero_delay_job",
		TargetURL:    ts.URL,
		Status:       "active",
		MaxRetries:   3,
		RetryDelayMs: 0,
	}

	_ = sched.AddJob(job)
	_ = sched.TriggerJob("zero_delay_job")

	time.Sleep(150 * time.Millisecond)

	if atomic.LoadInt32(&attempts) != 4 { // 1 initial + 3 retries
		t.Errorf("expected 4 attempts with zero delay, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestE2E_CR_G2_Tier2_NegativeBackoffMult_Defaults(t *testing.T) {
	// Negative or zero backoffMult should default to 1.0
	delay := cron.CalculateRetryDelay(100, -2.5, 2)
	if delay < 80*time.Millisecond || delay > 120*time.Millisecond {
		t.Errorf("negative backoffMult should default to 1.0, got delay %v", delay)
	}
}

func TestE2E_CR_G2_Tier2_RetryCount_TracksAttempts(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)
	job := &cron.Job{
		ID:           "track_retry_cnt",
		TargetURL:    ts.URL,
		Status:       "active",
		MaxRetries:   2,
		RetryDelayMs: 5,
	}

	_ = sched.AddJob(job)
	_ = sched.TriggerJob("track_retry_cnt")

	time.Sleep(200 * time.Millisecond)

	jobs := sched.GetJobs()
	for _, j := range jobs {
		if j.ID == "track_retry_cnt" {
			if j.RetryCount != 2 {
				t.Errorf("expected RetryCount 2, got %d", j.RetryCount)
			}
			if j.LastRetryAt.IsZero() {
				t.Errorf("expected LastRetryAt to be populated")
			}
		}
	}
}

func TestE2E_CR_G2_Tier2_HighMaxRetries_Cap(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	sched := cron.NewScheduler(nil)
	job := &cron.Job{
		ID:           "high_retry_job",
		TargetURL:    ts.URL,
		Status:       "active",
		MaxRetries:   5,
		RetryDelayMs: 1,
	}

	_ = sched.AddJob(job)
	_ = sched.TriggerJob("high_retry_job")

	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt32(&attempts) != 6 { // 1 + 5 retries = 6
		t.Errorf("expected 6 total attempts for MaxRetries=5, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestE2E_CR_G2_Tier2_NetworkTimeout_Retry(t *testing.T) {
	// Target URL points to closed port causing connection refused / timeout
	sched := cron.NewScheduler(nil)
	job := &cron.Job{
		ID:           "network_err_job",
		TargetURL:    "http://127.0.0.1:59999/nonexistent",
		Status:       "active",
		MaxRetries:   2,
		RetryDelayMs: 1,
	}

	_ = sched.AddJob(job)
	_ = sched.TriggerJob("network_err_job")

	time.Sleep(300 * time.Millisecond)

	jobs := sched.GetJobs()
	for _, j := range jobs {
		if j.ID == "network_err_job" {
			if j.LastOutcome != "failed" {
				t.Errorf("expected LastOutcome failed on network error, got %q", j.LastOutcome)
			}
			if j.FailureCount != 1 {
				t.Errorf("expected FailureCount 1, got %d", j.FailureCount)
			}
		}
	}
}

// ============================================================================
// CR.G4: Declarative YAML Cron-as-Code Definitions Tests
// ============================================================================

// --- Tier 1: Feature Coverage (CR.G4) ---

func TestE2E_CR_G4_Tier1_ParseYAML_SingleJob(t *testing.T) {
	yamlData := []byte(`
jobs:
  - id: daily-report
    cron: "0 9 * * 1-5"
    target_url: http://api/report
    on_success: notify-slack
`)

	jobs, err := config.ParseJobsYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseJobsYAML failed: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	j := jobs[0]
	if j.ID != "daily-report" {
		t.Errorf("expected ID daily-report, got %q", j.ID)
	}
	if j.Cron != "0 9 * * 1-5" {
		t.Errorf("expected Cron '0 9 * * 1-5', got %q", j.Cron)
	}
	if j.TargetURL != "http://api/report" {
		t.Errorf("expected TargetURL 'http://api/report', got %q", j.TargetURL)
	}
	if j.OnSuccess != "notify-slack" {
		t.Errorf("expected OnSuccess notify-slack, got %q", j.OnSuccess)
	}
}

func TestE2E_CR_G4_Tier1_ParseYAML_MultipleJobs(t *testing.T) {
	yamlData := []byte(`
jobs:
  - id: job-1
    target_url: http://api/1
    on_success: job-2
  - id: job-2
    target_url: http://api/2
    max_retries: 3
    retry_delay_ms: 500
    retry_backoff_mult: 2.0
`)

	jobs, err := config.ParseJobsYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseJobsYAML failed: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	if jobs[0].ID != "job-1" || jobs[0].OnSuccess != "job-2" {
		t.Errorf("job 1 mapping error: %+v", jobs[0])
	}
	if jobs[1].ID != "job-2" || jobs[1].MaxRetries != 3 || jobs[1].RetryDelayMs != 500 || jobs[1].RetryBackoffMult != 2.0 {
		t.Errorf("job 2 mapping error: %+v", jobs[1])
	}
}

func TestE2E_CR_G4_Tier1_LoadJobsFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "jobs.yaml")

	content := `
jobs:
  - id: file-job-a
    target_url: http://service/a
  - id: file-job-b
    target_url: http://service/b
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	jobs, err := config.LoadJobsFile(filePath)
	if err != nil {
		t.Fatalf("LoadJobsFile failed: %v", err)
	}

	if len(jobs) != 2 || jobs[0].ID != "file-job-a" || jobs[1].ID != "file-job-b" {
		t.Errorf("loaded jobs mismatch: %+v", jobs)
	}
}

func TestE2E_CR_G4_Tier1_WatchJobsFile_DetectsChange(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "watch_jobs.yaml")

	initialContent := `
jobs:
  - id: v1-job
    target_url: http://api/v1
`
	if err := os.WriteFile(filePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	var updatedJobs []cron.Job
	var mu sync.Mutex
	stopChan := make(chan struct{})
	defer close(stopChan)

	err := config.WatchJobsFileWithInterval(filePath, 50*time.Millisecond, func(jobs []cron.Job) {
		mu.Lock()
		updatedJobs = jobs
		mu.Unlock()
	}, stopChan)

	if err != nil {
		t.Fatalf("WatchJobsFileWithInterval failed: %v", err)
	}

	time.Sleep(70 * time.Millisecond)

	// Write updated file
	updatedContent := `
jobs:
  - id: v2-job
    target_url: http://api/v2
`
	_ = os.WriteFile(filePath, []byte(updatedContent), 0644)

	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(updatedJobs) != 1 || updatedJobs[0].ID != "v2-job" {
		t.Errorf("watch callback did not receive updated jobs: %+v", updatedJobs)
	}
}

func TestE2E_CR_G4_Tier1_FieldMapping_AllAttributes(t *testing.T) {
	yamlData := []byte(`
jobs:
  - id: full-job
    interval: 30s
    cron: "0 * * * *"
    target_url: http://api/full
    payload: '{"key":"val"}'
    next_topic: events
    status: active
    on_success: success-job
    on_failure: fail-job
    max_retries: 5
    retry_delay_ms: 1000
    retry_backoff_mult: 1.5
`)

	jobs, err := config.ParseJobsYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseJobsYAML failed: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	j := jobs[0]
	if j.ID != "full-job" || j.Interval != "30s" || j.Cron != "0 * * * *" || j.TargetURL != "http://api/full" {
		t.Errorf("basic attributes mismatch: %+v", j)
	}
	if j.Payload != `{"key":"val"}` || j.NextTopic != "events" || j.Status != "active" {
		t.Errorf("payload/topic/status mismatch: %+v", j)
	}
	if j.OnSuccess != "success-job" || j.OnFailure != "fail-job" {
		t.Errorf("DAG attributes mismatch: %+v", j)
	}
	if j.MaxRetries != 5 || j.RetryDelayMs != 1000 || j.RetryBackoffMult != 1.5 {
		t.Errorf("retry attributes mismatch: %+v", j)
	}
}

// --- Tier 2: Boundary & Corner Cases (CR.G4) ---

func TestE2E_CR_G4_Tier2_LoadJobsFile_NonExistentFile(t *testing.T) {
	_, err := config.LoadJobsFile("/path/does/not/exist/jobs.yaml")
	if err == nil {
		t.Errorf("expected error when loading non-existent file")
	}
}

func TestE2E_CR_G4_Tier2_ParseYAML_EmptyOrCommentsOnly(t *testing.T) {
	yamlData := []byte(`
# This is a comment file
# No jobs defined here
`)

	jobs, err := config.ParseJobsYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseJobsYAML failed: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs for comment-only YAML, got %d", len(jobs))
	}
}

func TestE2E_CR_G4_Tier2_ParseYAML_QuotedAndUnquotedStrings(t *testing.T) {
	yamlData := []byte(`
jobs:
  - id: "double-quoted"
    target_url: 'http://single-quoted'
  - id: unquoted
    target_url: http://unquoted
`)

	jobs, err := config.ParseJobsYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseJobsYAML failed: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0].ID != "double-quoted" || jobs[0].TargetURL != "http://single-quoted" {
		t.Errorf("quoted values mismatch: %+v", jobs[0])
	}
	if jobs[1].ID != "unquoted" || jobs[1].TargetURL != "http://unquoted" {
		t.Errorf("unquoted values mismatch: %+v", jobs[1])
	}
}

func TestE2E_CR_G4_Tier2_ParseYAML_InlineComments(t *testing.T) {
	yamlData := []byte(`
jobs:
  - id: comment-job # inline comment here
    target_url: http://api/test # main endpoint URL
    max_retries: 3 # retry limit
`)

	jobs, err := config.ParseJobsYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseJobsYAML failed: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].ID != "comment-job" || jobs[0].TargetURL != "http://api/test" || jobs[0].MaxRetries != 3 {
		t.Errorf("inline comment parsing failed: %+v", jobs[0])
	}
}

func TestE2E_CR_G4_Tier2_WatchJobsFile_StopChan(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "stop_watch.yaml")
	_ = os.WriteFile(filePath, []byte("jobs:\n  - id: j1\n    target_url: http://x\n"), 0644)

	stopChan := make(chan struct{})
	callCount := 0
	var mu sync.Mutex

	err := config.WatchJobsFileWithInterval(filePath, 20*time.Millisecond, func(jobs []cron.Job) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}, stopChan)

	if err != nil {
		t.Fatalf("WatchJobsFile failed: %v", err)
	}

	close(stopChan) // stop watcher
	time.Sleep(50 * time.Millisecond)

	// Update file after closing stopChan
	_ = os.WriteFile(filePath, []byte("jobs:\n  - id: j2\n    target_url: http://x\n"), 0644)
	time.Sleep(60 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if callCount != 0 {
		t.Errorf("watcher should not call onChange after stopChan closed, got callCount=%d", callCount)
	}
}

// ============================================================================
// Tier 3: Cross-Feature Combinations (CR.G1 + CR.G2 + CR.G4)
// ============================================================================

func TestE2E_CR_Tier3_Cross_YAMLLoadedDAGWithRetries(t *testing.T) {
	// Combination: Parse a full DAG pipeline definition from YAML containing retry policies,
	// register jobs into Scheduler, and execute.
	var executed []string
	var attemptsJobA int32
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		mu.Lock()
		executed = append(executed, id)
		mu.Unlock()

		if id == "etl-extract" {
			cnt := atomic.AddInt32(&attemptsJobA, 1)
			if cnt == 1 {
				w.WriteHeader(http.StatusGatewayTimeout) // fail 1st attempt
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	yamlSpec := fmt.Sprintf(`
jobs:
  - id: etl-extract
    target_url: "%s?id=etl-extract"
    on_success: etl-transform
    max_retries: 2
    retry_delay_ms: 10
  - id: etl-transform
    target_url: "%s?id=etl-transform"
    status: active
`, ts.URL, ts.URL)

	parsedJobs, err := config.ParseJobsYAML([]byte(yamlSpec))
	if err != nil {
		t.Fatalf("failed to parse YAML spec: %v", err)
	}

	sched := cron.NewScheduler(nil)
	for i := range parsedJobs {
		_ = sched.AddJob(&parsedJobs[i])
	}

	_ = sched.TriggerJob("etl-extract")
	time.Sleep(400 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// etl-extract fails 1st time, retries, succeeds 2nd time, then triggers etl-transform via OnSuccess
	if len(executed) != 3 || executed[0] != "etl-extract" || executed[1] != "etl-extract" || executed[2] != "etl-transform" {
		t.Errorf("expected [etl-extract, etl-extract, etl-transform], got %v", executed)
	}
}

// ============================================================================
// Tier 4: Real-World Application Scenarios (ServCron)
// ============================================================================

func TestE2E_CR_Tier4_Scenario_EtlDataPipeline(t *testing.T) {
	// Scenario: End-to-End Enterprise Data Pipeline
	// 1. Job "ingest-db": Connects to source DB (simulated HTTP endpoint). Fails once, retries automatically, then succeeds.
	// 2. OnSuccess -> "clean-data": Processes dataset and succeeds.
	// 3. OnSuccess -> "generate-report": Generates final PDF report.
	// 4. Fallback: If "clean-data" failed, OnFailure would trigger "alert-slack".

	var pipelineLog []string
	var mu sync.Mutex
	var ingestAttempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		step := r.URL.Query().Get("step")
		mu.Lock()
		pipelineLog = append(pipelineLog, step)
		mu.Unlock()

		if step == "ingest" {
			if atomic.AddInt32(&ingestAttempts, 1) == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	yamlPipeline := fmt.Sprintf(`
jobs:
  - id: step-1-ingest
    target_url: "%s?step=ingest"
    on_success: step-2-clean
    max_retries: 3
    retry_delay_ms: 15
    retry_backoff_mult: 1.2
  - id: step-2-clean
    target_url: "%s?step=clean"
    on_success: step-3-report
    on_failure: step-alert-slack
    status: active
  - id: step-3-report
    target_url: "%s?step=report"
    status: active
  - id: step-alert-slack
    target_url: "%s?step=alert"
    status: active
`, ts.URL, ts.URL, ts.URL, ts.URL)

	jobs, err := config.ParseJobsYAML([]byte(yamlPipeline))
	if err != nil {
		t.Fatalf("pipeline YAML parse error: %v", err)
	}

	sched := cron.NewScheduler(nil)
	for i := range jobs {
		if err := sched.AddJob(&jobs[i]); err != nil {
			t.Fatalf("failed to add job %s: %v", jobs[i].ID, err)
		}
	}

	// Trigger pipeline
	if err := sched.TriggerJob("step-1-ingest"); err != nil {
		t.Fatalf("failed to trigger pipeline start: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	expectedOrder := []string{"ingest", "ingest", "clean", "report"}
	if len(pipelineLog) != 4 {
		t.Fatalf("expected 4 steps in pipeline execution, got %d (%v)", len(pipelineLog), pipelineLog)
	}
	for i, step := range expectedOrder {
		if pipelineLog[i] != step {
			t.Errorf("step %d mismatch: expected %s, got %s", i, step, pipelineLog[i])
		}
	}
}
