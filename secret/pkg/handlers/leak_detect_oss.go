//go:build !enterprise

package import "net/http"

func LeakDetectionMiddleware(next http.Handler) http.Handler {
	// Open-source version does not scan for outbound secret leaks
	return next
}
