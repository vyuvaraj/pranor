package main

import (
	"os"
	"strings"
	"testing"
)

func TestAIScaffoldWrite(t *testing.T) {
	// Clean up environment and file targets
	os.Remove("main.pnr")

	// Set mock key so runtime does not log warn / fail on complete stub checks
	os.Setenv("OPENAI_API_KEY", "mock-key")
	os.Setenv("PRANOR_AI_CONNECTION", "openai://gpt-4o-mini")
	os.Setenv("PRANOR_TEST_AI_RESPONSE", "server \"8080\"\nroute \"GET\" \"/\" (req) { return \"ok\" }")

	// Run scaffold command mock
	runAIScaffold("Create an API serving user details", false)

	// Verify main.pnr was written
	if _, err := os.Stat("main.pnr"); os.IsNotExist(err) {
		t.Fatalf("expected main.pnr to be written by AI scaffolder")
	}

	content, err := os.ReadFile("main.pnr")
	if err != nil {
		t.Fatalf("failed to read generated main.pnr: %v", err)
	}

	if len(content) == 0 {
		t.Errorf("scaffolded file should not be empty")
	}

	// Clean up
	os.Remove("main.pnr")
}

func TestAIScaffoldAutoFix(t *testing.T) {
	// Clean up target
	os.Remove("main.pnr")

	// Set mock environment
	os.Setenv("OPENAI_API_KEY", "mock-key")
	os.Setenv("PRANOR_AI_CONNECTION", "openai://gpt-4o-mini")

	// Response 1: Valid Serv syntax but fails unit tests (assert x == 2)
	// Response 2: Corrected code passing the tests
	resp1 := `server "8080"
test "failing test" {
	let x = 1
	assert x == 2
}`
	resp2 := `server "8080"
test "passing test" {
	let x = 1
	assert x == 1
}`

	os.Setenv("PRANOR_TEST_AI_RESPONSE", resp1+"|||"+resp2)

	// Run with autoFix = true
	runAIScaffold("create API and write tests", true)

	// Verify main.pnr was written and contains correct content from response 2
	if _, err := os.Stat("main.pnr"); os.IsNotExist(err) {
		t.Fatalf("expected main.pnr to be written by AI scaffolder auto-fix")
	}

	content, err := os.ReadFile("main.pnr")
	if err != nil {
		t.Fatalf("failed to read main.pnr: %v", err)
	}

	if !strings.Contains(string(content), "assert x == 1") {
		t.Errorf("expected generated file to contain the fixed assertion, got:\n%s", string(content))
	}

	// Clean up
	os.Remove("main.pnr")
	os.Unsetenv("PRANOR_TEST_AI_RESPONSE")
}
