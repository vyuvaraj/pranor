package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const defaultAdminURL = "http://localhost:8081"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "status":
		handleStatus()
	case "routes":
		handleRoutesCommand(os.Args[2:])
	case "version":
		fmt.Println("servgateway CLI v2.0.0")
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: servgateway <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  status                Check status of running servgatewayd daemon")
	fmt.Println("  routes list           List active gateway proxy routes")
	fmt.Println("  routes add <path> <target_url>  Add new proxy route to running daemon")
	fmt.Println("  version               Show CLI version")
}

func handleStatus() {
	resp, err := http.Get(defaultAdminURL + "/api/v1/health")
	if err != nil {
		fmt.Printf("Error connecting to servgatewayd daemon at %s: %v\n", defaultAdminURL, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var statusMap map[string]interface{}
	_ = json.Unmarshal(body, &statusMap)

	fmt.Println("ServGateway Daemon Status:")
	for k, v := range statusMap {
		fmt.Printf("  %-15s: %v\n", k, v)
	}
}

func handleRoutesCommand(args []string) {
	if len(args) == 0 || args[0] == "list" {
		resp, err := http.Get(defaultAdminURL + "/api/v1/routes")
		if err != nil {
			fmt.Printf("Error fetching routes: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Active Gateway Routes:")
		fmt.Println(string(body))
		return
	}

	if args[0] == "add" && len(args) >= 3 {
		path := args[1]
		target := args[2]

		payload, _ := json.Marshal(map[string]string{
			"path":       path,
			"target_url": target,
		})

		resp, err := http.Post(defaultAdminURL+"/api/v1/routes", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			fmt.Printf("Error adding route: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			fmt.Printf("Successfully registered route: %s -> %s\n", path, target)
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Failed to add route: %s\n", string(body))
		}
		return
	}

	fmt.Println("Invalid routes command. Use 'servgateway routes list' or 'servgateway routes add <path> <target_url>'")
}
