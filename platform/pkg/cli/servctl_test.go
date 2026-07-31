package cli

import (
	"strings"
	"testing"
)

func TestServctlCLI_Execute(t *testing.T) {
	cli := NewServctlCLI()

	// get services
	res := cli.Execute("get services")
	if !res.Success {
		t.Fatalf("expected 'get services' to succeed, got: %s", res.Error)
	}

	// restart service with name
	res = cli.Execute("restart service payment-api")
	if !res.Success || !strings.Contains(res.Output.(string), "payment-api") {
		t.Errorf("expected restart to succeed with service name, got: %+v", res)
	}

	// restart service without name
	res = cli.Execute("restart service")
	if res.Success {
		t.Error("expected restart service without name to fail")
	}

	// unknown command
	res = cli.Execute("nuke everything")
	if res.Success {
		t.Error("expected unknown command to fail")
	}

	// get nodes
	res = cli.Execute("get nodes")
	if !res.Success {
		t.Fatalf("expected 'get nodes' to succeed, got: %s", res.Error)
	}
}
