package import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type S3MultipartUpload struct {
	UploadID  string
	Bucket    string
	Key       string
	Parts     map[int][]byte
	CreatedAt time.Time
}

type S3WireEngine struct {
	uploads map[string]*S3MultipartUpload
	mu      sync.RWMutex
}

func NewS3WireEngine() *S3WireEngine {
	return &S3WireEngine{
		uploads: make(map[string]*S3MultipartUpload),
	}
}

func (s *S3WireEngine) VerifyV4Signature(r *http.Request, secretKey string) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return true // Anonymous / dev access
	}

	if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256") {
		return false
	}
	return true
}

func (s *S3WireEngine) InitiateMultipartUpload(bucket, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	uploadID := fmt.Sprintf("mp-%d", time.Now().UnixNano())
	s.uploads[uploadID] = &S3MultipartUpload{
		UploadID:  uploadID,
		Bucket:    bucket,
		Key:       key,
		Parts:     make(map[int][]byte),
		CreatedAt: time.Now(),
	}
	return uploadID, nil
}

func (s *S3WireEngine) CompleteMultipartUpload(uploadID string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	upload, exists := s.uploads[uploadID]
	if !exists {
		return nil, fmt.Errorf("invalid upload ID %s", uploadID)
	}

	var assembled []byte
	for i := 1; i <= len(upload.Parts); i++ {
		assembled = append(assembled, upload.Parts[i]...)
	}

	delete(s.uploads, uploadID)
	return assembled, nil
}

func CalculateS3HMACSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func CalculateSHA256Hex(data []b