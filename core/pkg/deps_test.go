package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceDependenciesAndModules(t *testing.T) {
	// Root directory is 2 levels up from packages/Pranor Core/pkg
	rootDir, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("failed to locate root dir: %v", err)
	}

	goWorkPath := filepath.Join(rootDir, "go.work")
	data, err := os.ReadFile(goWorkPath)
	if err != nil {
		t.Fatalf("failed to read go.work at %s: %v", goWorkPath, err)
	}

	content := string(data)
	expectedPackages := []string{
		"packages/Pranor",
		"packages/Pranor Gate",
		"packages/Pranor Pulse",
		"packages/Pranor Vault",
		"packages/Pranor Cache",
		"packages/Pranor Auth",
		"packages/Pranor Console",
		"packages/Pranor Mesh",
		"packages/Pranor Chrono",
		"packages/Pranor Deploy",
		"packages/Pranor Trace",
		"packages/Pranor Tunnel",
		"packages/Pranor Pool",
		"packages/Pranor Notify",
		"packages/Pranor Flow",
		"packages/Pranor Hub",
		"packages/Pranor Core",
		"packages/servlockctl",
		"packages/servsecretctl",
	}

	for _, pkg := range expectedPackages {
		if !strings.Contains(content, pkg) {
			t.Errorf("expected go.work to register package %q", pkg)
		}

		goModFile := filepath.Join(rootDir, pkg, "go.mod")
		if _, err := os.Stat(goModFile); os.IsNotExist(err) {
			t.Errorf("expected go.mod to exist at %s", goModFile)
		}
	}
}
