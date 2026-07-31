package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceDependenciesAndModules(t *testing.T) {
	// Root directory is 2 levels up from core/pkg (i.e. pranor/)
	rootDir, err := filepath.Abs(filepath.Join("..", ".."))
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
		"./lang",
		"./gate",
		"./pulse",
		"./vault",
		"./cache",
		"./auth",
		"./console",
		"./mesh",
		"./chrono",
		"./deploy",
		"./trace",
		"./tunnel",
		"./pool",
		"./notify",
		"./flow",
		"./hub",
		"./core",
		"./lockctl",
		"./secretctl",
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
