package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ServQueue Standalone CLI (servqueue)")
		fmt.Println("Usage: servqueue <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  status               Check cluster node health status")
		fmt.Println("  publish              Publish an event payload to a topic")
		fmt.Println("  consume              Read events from a topic")
		os.Exit(0)
	}

	cmd := os.Args[1]
	serverURL := os.Getenv("SERVQUEUE_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:9092"
	}

	switch cmd {
	case "status":
		resp, err := http.Get(serverURL + "/health")
		if err != nil {
			fmt.Printf("Error connecting to servqueued at %s: %v\n", serverURL, err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("ServQueue Node Status: %s\n", string(body))

	case "publish":
		pubCmd := flag.NewFlagSet("publish", flag.ExitOnError)
		topic := pubCmd.String("topic", "", "Topic name")
		payload := pubCmd.String("payload", "", "Event payload JSON string")
		_ = pubCmd.Parse(os.Args[2:])

		if *topic == "" || *payload == "" {
			fmt.Println("Usage: servqueue publish --topic <name> --payload '<json>'")
			os.Exit(1)
		}

		reqBody, _ := json.Marshal(map[string]string{"topic": *topic, "payload": *payload})
		resp, err := http.Post(serverURL+"/api/enqueue", "application/json", bytes.NewReader(reqBody))
		if err != nil {
			fmt.Printf("Publish failed: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		fmt.Printf("Event successfully published to topic '%s'\n", *topic)

	default:
		fmt.Printf("Unknown command '%s'. Run 'servqueue' for usage.\n", cmd)
	}
}
