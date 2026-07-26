package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const defaultAdminURL = "http://localhost:9001"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "status":
		handleStatus()
	case "ls":
		handleListBuckets()
	case "mb":
		if len(os.Args) < 3 {
			fmt.Println("Usage: servstore mb <bucket_name>")
			os.Exit(1)
		}
		handleMakeBucket(os.Args[2])
	case "version":
		fmt.Println("servstore CLI v2.0.0")
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: servstore <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  status               Check status of running servstored daemon")
	fmt.Println("  ls                   List active storage buckets")
	fmt.Println("  mb <bucket_name>     Create a new storage bucket")
	fmt.Println("  version              Show CLI version")
}

func handleStatus() {
	resp, err := http.Get(defaultAdminURL + "/api/v1/health")
	if err != nil {
		fmt.Printf("Error connecting to servstored daemon at %s: %v\n", defaultAdminURL, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var statusMap map[string]interface{}
	_ = json.Unmarshal(body, &statusMap)

	fmt.Println("ServStore Daemon Status:")
	for k, v := range statusMap {
		fmt.Printf("  %-15s: %v\n", k, v)
	}
}

func handleListBuckets() {
	resp, err := http.Get(defaultAdminURL + "/api/v1/buckets")
	if err != nil {
		fmt.Printf("Error listing buckets: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Active Storage Buckets:")
	fmt.Println(string(body))
}

func handleMakeBucket(bucketName string) {
	payload, _ := json.Marshal(map[string]string{
		"name": bucketName,
	})

	resp, err := http.Post(defaultAdminURL+"/api/v1/buckets", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error creating bucket: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		fmt.Printf("Successfully created bucket: %s\n", bucketName)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Failed to create bucket: %s\n", string(body))
	}
}
