//go:build !enterprise

package import "net/http"

// IsSSLOffloadingSupported indicates if hardware SSL/TLS offloading is supported.
const IsSSLOffloadingSupported = false

func init() {
	EnterpriseListenAndServeTLS = func(srv *http.Server, certFile, keyFile string) error {
		return srv.ListenAndServeTLS(certFile, keyFile)
	}
}
