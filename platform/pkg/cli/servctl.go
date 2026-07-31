package import (
	"encoding/json"
	"fmt"
	"strings"
)

// ServctlCommand represents a parsed servctl CLI command.
type ServctlCommand struct {
	Verb       string            `json:"verb"`       // e.g. "get", "apply", "delete", "restart"
	Resource   string            `json:"resource"`   // e.g. "service", "queue", "gateway"
	Name       string            `json:"name"`       // resource name or wildcard
	Flags      map[string]string `json:"flags"`
}

// ServctlResult represents execution result of a CLI command.
type ServctlResult struct {
	Command string      `json:"command"`
	Success bool        `json:"success"`
	Output  interface{} `json:"output,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ServctlCLI is the cluster-wide administration CLI for the Pranor platform.
type ServctlCLI struct {
	registry map[string]func(cmd ServctlCommand) ServctlResult
}

// NewServctlCLI creates a ServctlCLI instance with built-in command handlers.
func NewServctlCLI() *ServctlCLI {
	cli := &ServctlCLI{
		registry: make(map[string]func(cmd ServctlCommand) ServctlResult),
	}
	cli.registerBuiltins()
	return cli
}

func (sc *ServctlCLI) registerBuiltins() {
	sc.registry["get services"] = func(cmd ServctlCommand) ServctlResult {
		return ServctlResult{Command: "get services", Success: true, Output: []string{"Pranor Gate", "Pranor Pulse", "Pranor Vault", "Pranor Mesh", "Pranor Trace"}}
	}
	sc.registry["restart service"] = func(cmd ServctlCommand) ServctlResult {
		if cmd.Name == "" {
			return ServctlResult{Command: "restart service", Success: false, Error: "service name is required"}
		}
		return ServctlResult{Command: fmt.Sprintf("restart service %s", cmd.Name), Success: true, Output: fmt.Sprintf("service '%s' restarted successfully", cmd.Name)}
	}
	sc.registry["get nodes"] = func(cmd ServctlCommand) ServctlResult {
		return ServctlResult{Command: "get nodes", Success: true, Output: []string{"node-1", "node-2", "node-3"}}
	}
	sc.registry["apply config"] = func(cmd ServctlCommand) ServctlResult {
		return ServctlResult{Command: "apply config", Success: true, Output: "configuration applied to cluster"}
	}
}

// Execute parses and runs a servctl CLI command string.
func (sc *ServctlCLI) Execute(rawCommand string) ServctlResult {
	parts := strings.Fields(strings.TrimSpace(rawCommand))
	if len(parts) < 2 {
		return ServctlResult{Command: rawCommand, Success: false, Error: "usage: servctl <verb> <resource> [name]"}
	}

	verb := parts[0]
	resource := parts[1]
	name := ""
	if len(parts) > 2 {
		name = parts[2]
	}

	flags := make(map[string]string)
	for i := 3; i < len(parts); i++ {
		if strings.HasPrefix(parts[i], "--") {
			kv := strings.SplitN(strings.TrimPrefix(parts[i], "--"), "=", 2)
			if len(kv) == 2 {
				flags[kv[0]] = kv[1]
			}
		}
	}

	cmd := ServctlCommand{Verb: verb, Resource: resource, Name: name, Flags: flags}
	key := fmt.Sprintf("%s %s", verb, resource)

	handler, ok := sc.registry[key]
	if !ok {
		return ServctlResult{Command: rawCommand, Success: false, Error: fmt.Sprintf("unknown command: %s", key)}
	}
	return handler(cmd)
}

// ExecuteJSON runs a command and returns result as JSON string.
func (sc *ServctlCLI) ExecuteJSON(rawCommand string) string {
	result := sc.Execute(rawCommand)
	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b)
}
