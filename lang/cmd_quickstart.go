package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func runQuickstart() {
	fmt.Println("🚀 Pranor Interactive CLI Quickstart Wizard")
	fmt.Println("===========================================")
	fmt.Println("Welcome! Let's scaffold your infrastructure & application in 3 quick steps.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("1. What project name would you like to use? [default: my-pranor-app]: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = "my-pranor-app"
	}

	fmt.Println("\n2. Select required modules/components:")
	fmt.Println("   [1] REST API + Ingress Gateway (Pranor Gate)")
	fmt.Println("   [2] Event Queue & Messaging (Pranor Pulse)")
	fmt.Println("   [3] Object Storage & Database (Pranor Store)")
	fmt.Println("   [4] Full Stack Monolith (All modules)")
	fmt.Print("Choice [1-4, default: 4]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	template := "full-stack"
	switch choice {
	case "1":
		template = "api"
	case "2":
		template = "worker"
	case "3":
		template = "event-processor"
	default:
		template = "full-stack"
	}

	fmt.Printf("\n🔨 Scaffolding project %q with template %q...\n", name, template)
	createNewProject(name, template)

	fmt.Println("\n✅ Quickstart complete!")
	fmt.Printf("   To get started:\n")
	fmt.Printf("     cd %s\n", name)
	fmt.Printf("     pranor run main.pnr\n\n")
}
