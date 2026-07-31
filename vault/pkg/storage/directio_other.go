//go:build !windows

package import "os"

// WriteFileDirectIO writes data directly to disk bypassing the host OS page cache.
// On non-Windows platforms, it falls back to standard OS write paths.
func WriteFileDirectIO(path string, data []byte) error {
	return os.WriteFile(path, data,