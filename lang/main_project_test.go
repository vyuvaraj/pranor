package lang

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryMerge(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "PRANOR_dir_merge_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple .pnr files that reference each other without imports
	fileAContent := `
export fn getValue() -> int {
	return 42
}
`
	fileBContent := `
fn useValue() -> int {
	return getValue()
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "a.pnr"), []byte(fileAContent), 0644); err != nil {
		t.Fatalf("failed to write a.pnr: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.pnr"), []byte(fileBContent), 0644); err != nil {
		t.Fatalf("failed to write b.pnr: %v", err)
	}

	// Compile the directory
	outBin := "test_service.exe"
	binPath, err := buildServNoExit(tmpDir, outBin, "", "", "", "")
	if err != nil {
		t.Fatalf("failed to build directory: %v", err)
	}
	defer os.Remove(binPath)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Errorf("expected binary to be generated at %s", binPath)
	}
}

func TestManifestBuild(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "PRANOR_manifest_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create pranor.toml pointing to a custom entry file
	tomlContent := `
name = "my-test-project"
version = "1.0.0"
entry = "custom_entry.pnr"
`
	entryContent := `
fn someFunc() -> int {
	return 100
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "pranor.toml"), []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write pranor.toml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "custom_entry.pnr"), []byte(entryContent), 0644); err != nil {
		t.Fatalf("failed to write custom_entry.pnr: %v", err)
	}

	// Compile the directory
	outBin := "test_service_manifest.exe"
	binPath, err := buildServNoExit(tmpDir, outBin, "", "", "", "")
	if err != nil {
		t.Fatalf("failed to build directory via manifest: %v", err)
	}
	defer os.Remove(binPath)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Errorf("expected binary to be generated at %s", binPath)
	}
}

func TestNewAndDeployK8s(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "PRANOR_new_deploy_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	projDir := filepath.Join(tmpDir, "test-api-proj")

	// 1. Test template creation (api template)
	createNewProject(projDir, "api")

	// Verify scaffolded files
	if _, err := os.Stat(filepath.Join(projDir, "main.pnr")); os.IsNotExist(err) {
		t.Error("expected main.pnr to be created")
	}
	if _, err := os.Stat(filepath.Join(projDir, "config.yml")); os.IsNotExist(err) {
		t.Error("expected config.yml to be created")
	}
	if _, err := os.Stat(filepath.Join(projDir, "main_test.pnr")); os.IsNotExist(err) {
		t.Error("expected main_test.pnr to be created")
	}

	// 1.1 Test other templates
	templates := []string{"worker", "event-processor", "full-stack"}
	for _, templ := range templates {
		d := filepath.Join(tmpDir, "test-"+templ)
		createNewProject(d, templ)
		if _, err := os.Stat(filepath.Join(d, "main.pnr")); os.IsNotExist(err) {
			t.Errorf("expected main.pnr to be created for template %s", templ)
		}
		// Try parsing to verify syntax is valid
		_, _, err := parseProject(filepath.Join(d, "main.pnr"))
		if err != nil {
			t.Errorf("scaffolded project for template %s has invalid syntax: %v", templ, err)
		}
	}


	// 2. Test deploy command for k8s target
	deployServ(filepath.Join(projDir, "main.pnr"), "k8s")

	// Verify generated Kubernetes manifests and Dockerfile
	if _, err := os.Stat(filepath.Join(projDir, "Dockerfile")); os.IsNotExist(err) {
		t.Error("expected Dockerfile to be created")
	}
	if _, err := os.Stat(filepath.Join(projDir, "k8s", "deployment.yaml")); os.IsNotExist(err) {
		t.Error("expected k8s/deployment.yaml to be created")
	}
	if _, err := os.Stat(filepath.Join(projDir, "k8s", "service.yaml")); os.IsNotExist(err) {
		t.Error("expected k8s/service.yaml to be created")
	}
	if _, err := os.Stat(filepath.Join(projDir, "k8s", "configmap.yaml")); os.IsNotExist(err) {
		t.Error("expected k8s/configmap.yaml to be created")
	}

	// Read and verify deployment contains volumeMount
	depContent, err := os.ReadFile(filepath.Join(projDir, "k8s", "deployment.yaml"))
	if err != nil {
		t.Fatalf("failed to read deployment.yaml: %v", err)
	}
	if !testContains(string(depContent), "config-volume") {
		t.Errorf("expected deployment to mount config volume")
	}
}

func TestAICompilation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_ai_compilation_*.pnr")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	srvContent := `
ai "openai://gpt-4o-mini"
server "8080"

route "POST" "/ask" (req) {
	let res = ai.complete(req.body)
	let vec = ai.embed("text to embed")
	return { "res": res, "vector": vec }
}
`
	if _, err := tmpFile.WriteString(srvContent); err != nil {
		t.Fatalf("failed to write srv file: %v", err)
	}
	tmpFile.Close()

	binPath, err := buildServNoExit(tmpFile.Name(), "temp_ai_test.exe", "", "", "", "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	defer os.Remove(binPath)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Errorf("expected binary to be generated at %s", binPath)
	}
}

func testContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || stringsContains(s, substr))
}

func stringsContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
