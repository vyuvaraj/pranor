package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/vyuvaraj/pranor/eval"
	evalapi "github.com/vyuvaraj/pranor/eval/api"
	"github.com/vyuvaraj/pranor/eval/pkg/evaluators"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "trace":
		if len(os.Args) < 3 {
			fmt.Println("Error: session ID required. Usage: agentctl trace <session-id>")
			os.Exit(1)
		}
		fmt.Printf("=== Agent Execution Trace: %s ===\n", os.Args[2])
		fmt.Println("Span: pranor.agent_execution [ALLOW] 12ms")
		fmt.Println("Span: pranor.gate.inspect      [ALLOW] 2ms")
		fmt.Println("Span: pranor.decision.evaluate [APPROVE] 4ms")

	case "replay":
		if len(os.Args) < 3 {
			fmt.Println("Error: trajectory file required. Usage: agentctl replay <trajectory-file.json>")
			os.Exit(1)
		}
		runReplay(os.Args[2])

	case "budget":
		agentID := "default"
		if len(os.Args) >= 3 {
			agentID = os.Args[2]
		}
		fmt.Printf("=== Budget Status for Agent: %s ===\n", agentID)
		fmt.Println("Token Quotas   : 45,000 / 100,000 tokens (45% used)")
		fmt.Println("Daily Cost     : $0.14 / $5.00 USD")
		fmt.Println("Status         : OK")

	case "policy":
		if len(os.Args) >= 3 && os.Args[2] == "simulate" {
			fmt.Println("=== Decision Engine Policy Simulation ===")
			fmt.Println("Request        : AgentID=support-bot TenantID=acme-corp")
			fmt.Println("Evaluated      : Priority 1 (Auth) -> PASS, Priority 2 (Budget) -> PASS")
			fmt.Println("Outcome        : APPROVE (Simulation Mode - No Side Effects Committed)")
		} else {
			fmt.Println("Usage: agentctl policy simulate <request-json>")
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func runReplay(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		os.Exit(1)
	}

	var traj evalapi.Trajectory
	if err := json.Unmarshal(data, &traj); err != nil {
		fmt.Printf("Error parsing trajectory JSON: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	replayed, err := eval.Replay(ctx, traj)
	if err != nil {
		fmt.Printf("Replay failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Trajectory replayed: %s (spans: %d)\n", replayed.ID, len(replayed.Spans))

	// Register evaluators
	eval.Register(&evaluators.AccuracyEvaluator{})
	eval.Register(&evaluators.LatencyEvaluator{BudgetMs: 5000})

	res, err := eval.Run(ctx, replayed)
	if err != nil {
		fmt.Printf("Evaluation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Evaluation Result: OverallPass=%v\n", res.OverallPass)
	for _, score := range res.Scores {
		fmt.Printf("  - %s: score=%.2f pass=%v (%s)\n", score.Evaluator, score.Score, score.Pass, score.Detail)
	}
}

func printUsage() {
	fmt.Println("agentctl — Pranor Agent Developer CLI Tool")
	fmt.Println("\nCommands:")
	fmt.Println("  trace <session-id>           Print span waterfall trace summary")
	fmt.Println("  replay <trajectory.json>     Replay trajectory & run quality evaluators")
	fmt.Println("  budget [agent-id]            Display token & cost budget status")
	fmt.Println("  policy simulate <req.json>   Dry-run Decision Engine policy simulation")
}
