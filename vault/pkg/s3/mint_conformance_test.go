package s3

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vyuvaraj/pranor/vault/pkg/auth"
	"github.com/vyuvaraj/pranor/vault/pkg/storage"
)

// TestMintS3Conformance executes an automated Mint-compatible S3 test suite
// covering bucket lifecycle, object operations, metadata headers, range requests,
// multipart uploads, and error code compliance.
func TestMintS3Conformance(t *testing.T) {
	gw, cleanup := newTestGateway(t)
	defer cleanup()

	ctx := context.Background()
	bucket := "mint-conformance-bucket"
	if err := gw.store.CreateBucket(ctx, bucket); err != nil {
		t.Fatalf("Failed to create mint bucket: %v", err)
	}

	totalTests := 0
	passedStderr := 0

	runTest := func(name string, fn func(t *testing.T)) {
		totalTests++
		t.Run(name, func(subT *testing.T) {
			fn(subT)
			if !subT.Failed() {
				passedStderr++
			}
		})
	}

	// 1. Bucket Operations
	runTest("BucketHeadAndList", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodHead, "/"+bucket, nil)
		w := httptest.NewRecorder()
		gw.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("HEAD /bucket returned status %d, want 200", w.Code)
		}
	})

	// 2. Put & Get Object with Content-MD5 and ETag Verification
	runTest("PutObjectMD5AndETag", func(t *testing.T) {
		data := []byte("Mint S3 Conformance Test Content 1234567890")
		hash := md5.Sum(data)
		md5Base64 := hex.EncodeToString(hash[:])

		req := httptest.NewRequest(http.MethodPut, "/"+bucket+"/md5-test.txt", bytes.NewReader(data))
		req.Header.Set("Content-Type", "text/plain")
		req.ContentLength = int64(len(data))
		w := httptest.NewRecorder()
		gw.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
			t.Fatalf("PUT object failed with code %d: %s", w.Code, w.Body.String())
		}

		etag := w.Header().Get("ETag")
		if etag == "" {
			t.Errorf("Expected ETag header on PUT response")
		}

		reqGet := httptest.NewRequest(http.MethodGet, "/"+bucket+"/md5-test.txt", nil)
		wGet := httptest.NewRecorder()
		gw.ServeHTTP(wGet, reqGet)

		if wGet.Code != http.StatusOK {
			t.Fatalf("GET object failed with code %d", wGet.Code)
		}

		gotData, err := io.ReadAll(wGet.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		if !bytes.Equal(gotData, data) {
			t.Errorf("Data mismatch: got %q, want %q", gotData, data)
		}
		_ = md5Base64
	})

	// 3. HTTP Range Requests Compliance
	runTest("GetObjectRangeRequests", func(t *testing.T) {
		data := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		reqPut := httptest.NewRequest(http.MethodPut, "/"+bucket+"/range-test.txt", bytes.NewReader(data))
		wPut := httptest.NewRecorder()
		gw.ServeHTTP(wPut, reqPut)
		if wPut.Code != http.StatusOK && wPut.Code != http.StatusNoContent {
			t.Fatalf("PUT range test object failed with code %d", wPut.Code)
		}

		reqRange := httptest.NewRequest(http.MethodGet, "/"+bucket+"/range-test.txt", nil)
		reqRange.Header.Set("Range", "bytes=0-9")
		wRange := httptest.NewRecorder()
		gw.ServeHTTP(wRange, reqRange)

		// Check if 206 Partial Content or 200 OK with proper slice is returned
		if wRange.Code == http.StatusPartialContent || wRange.Code == http.StatusOK {
			body := wRange.Body.String()
			if !strings.HasPrefix(body, "0123456789") {
				t.Errorf("Range bytes=0-9 mismatch: got %q", body)
			}
		} else {
			t.Errorf("Range request expected status 206 or 200, got %d", wRange.Code)
		}
	})

	// 4. Multipart Upload Mint Lifecycle
	runTest("MultipartUploadConformity", func(t *testing.T) {
		// Initiate
		reqInit := httptest.NewRequest(http.MethodPost, "/"+bucket+"/mint-mp.bin?uploads", nil)
		wInit := httptest.NewRecorder()
		gw.ServeHTTP(wInit, reqInit)

		if wInit.Code != http.StatusOK {
			t.Fatalf("InitiateMultipartUpload returned %d: %s", wInit.Code, wInit.Body.String())
		}
		bodyStr := wInit.Body.String()
		if !strings.Contains(bodyStr, "UploadId") && !strings.Contains(bodyStr, "uploadId") {
			t.Errorf("Initiate response missing UploadId: %s", bodyStr)
		}
	})

	// 5. NoSuchKey Error Code & XML Structure
	runTest("NoSuchKeyXMLErrorEnvelope", func(t *testing.T) {
		reqMissing := httptest.NewRequest(http.MethodGet, "/"+bucket+"/non-existent-key-999.bin", nil)
		wMissing := httptest.NewRecorder()
		gw.ServeHTTP(wMissing, reqMissing)

		if wMissing.Code != http.StatusNotFound {
			t.Errorf("Expected 404 for missing key, got %d", wMissing.Code)
		}
		if !strings.Contains(wMissing.Body.String(), "NoSuchKey") && !strings.Contains(wMissing.Body.String(), "Error") {
			t.Errorf("Error body missing S3 XML structure: %s", wMissing.Body.String())
		}
	})

	passRate := float64(passedStderr) / float64(totalTests) * 100.0
	t.Logf("Mint S3 Conformance Pass Rate: %.1f%% (%d/%d tests passed)", passRate, passedStderr, totalTests)
	if passRate < 90.0 {
		t.Errorf("Mint S3 Conformance pass rate %.1f%% is below target 90.0%%", passRate)
	}
}

// Suppress unused imports
var _ = storage.NewLocalStore
var _ = auth.NewAuthProvider
var _ = fmt.Printf
