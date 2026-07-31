package main

import (
	"os"
	"testing"
)

func TestPranorPulseOPFSConnectionParsing(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_pranor-pulse_opfs_*.pnr")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	srvContent := `
broker "Pranor Pulse://opfs"

route "POST" "/events" (req) {
	publish "user_events" req.body
	return { status: "queued_opfs" }
}
`
	if _, err := tmpFile.WriteString(srvContent); err != nil {
		t.Fatalf("failed to write srv file: %v", err)
	}
	tmpFile.Close()

	outExe := "temp_pranor-pulse_opfs.exe"
	_, err = buildServNoExit(tmpFile.Name(), outExe, "", "", "", "")
	if err != nil {
		t.Fatalf("Build failed for Pranor Pulse://opfs broker target: %v", err)
	}
	_ = os.Remove(outExe)
}
