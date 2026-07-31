package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vyuvaraj/pranor/chrono/pkg/cron"
)

// LoadJobsFile parses a YAML file into []cron.Job.
func LoadJobsFile(path string) ([]cron.Job, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read jobs file %s: %w", path, err)
	}
	return ParseJobsYAML(data)
}

// WatchJobsFile monitors a jobs YAML file for changes using 5-second os.Stat polling.
func WatchJobsFile(path string, onChange func([]cron.Job)) error {
	return WatchJobsFileWithInterval(path, 5*time.Second, onChange, nil)
}

// WatchJobsFileWithInterval allows custom polling interval and optional stop channel for testing.
func WatchJobsFileWithInterval(path string, interval time.Duration, onChange func([]cron.Job), stopChan <-chan struct{}) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("initial stat failed for %s: %w", path, err)
	}

	lastModTime := stat.ModTime()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				st, err := os.Stat(path)
				if err != nil {
					continue
				}
				if st.ModTime().After(lastModTime) {
					lastModTime = st.ModTime()
					jobs, err := LoadJobsFile(path)
					if err == nil {
						onChange(jobs)
					}
				}
			}
		}
	}()

	return nil
}

// ParseJobsYAML parses a minimal YAML subset containing a "jobs" list into []cron.Job.
func ParseJobsYAML(data []byte) ([]cron.Job, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var jobs []cron.Job
	var current *cron.Job
	inJobs := false

	for scanner.Scan() {
		line := scanner.Text()

		// Strip inline comments starting with '#' preceded by a space
		if idx := strings.Index(line, " #"); idx != -1 {
			line = line[:idx]
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if trimmed == "jobs:" {
			inJobs = true
			continue
		}

		if !inJobs {
			continue
		}

		// Check if line introduces a new list item
		isListItem := false
		itemContent := trimmed
		if strings.HasPrefix(trimmed, "- ") {
			isListItem = true
			itemContent = strings.TrimPrefix(trimmed, "- ")
		} else if trimmed == "-" {
			isListItem = true
			itemContent = ""
		}

		if isListItem {
			if current != nil {
				jobs = append(jobs, *current)
			}
			current = &cron.Job{}
			if itemContent == "" {
				continue
			}
		}

		if current == nil {
			continue
		}

		// Split on first colon
		colonIdx := strings.Index(itemContent, ":")
		if colonIdx == -1 {
			continue
		}

		key := strings.TrimSpace(itemContent[:colonIdx])
		val := strings.TrimSpace(itemContent[colonIdx+1:])
		val = unquote(val)

		switch key {
		case "id":
			current.ID = val
		case "interval":
			current.Interval = val
		case "cron":
			current.Cron = val
		case "target_url":
			current.TargetURL = val
		case "payload":
			current.Payload = val
		case "next_topic":
			current.NextTopic = val
		case "status":
			current.Status = val
		case "on_success":
			current.OnSuccess = val
		case "on_failure":
			current.OnFailure = val
		case "max_retries":
			if v, err := strconv.Atoi(val); err == nil {
				current.MaxRetries = v
			}
		case "retry_delay_ms":
			if v, err := strconv.Atoi(val); err == nil {
				current.RetryDelayMs = v
			}
		case "retry_backoff_mult":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				current.RetryBackoffMult = v
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning YAML: %w", err)
	}

	if current != nil {
		jobs = append(jobs, *current)
	}

	return jobs, nil
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
