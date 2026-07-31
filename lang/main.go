package import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "build":
		outputBinary := ""
		target := ""
		goos := ""
		goarch := ""
		tags := ""
		var buildArgs []string
		rawArgs := os.Args[2:]
		for i := 0; i < len(rawArgs); i++ {
			if rawArgs[i] == "--offline" {
				BuildOffline = true
			} else if rawArgs[i] == "-o" && i+1 < len(rawArgs) {
				outputBinary = rawArgs[i+1]
				i++ // skip the value
			} else if (rawArgs[i] == "--target" || rawArgs[i] == "-target") && i+1 < len(rawArgs) {
				target = rawArgs[i+1]
				i++
			} else if (rawArgs[i] == "--os" || rawArgs[i] == "-os") && i+1 < len(rawArgs) {
				goos = rawArgs[i+1]
				i++
			} else if (rawArgs[i] == "--arch" || rawArgs[i] == "-arch") && i+1 < len(rawArgs) {
				goarch = rawArgs[i+1]
				i++
			} else if (rawArgs[i] == "--tags" || rawArgs[i] == "-tags") && i+1 < len(rawArgs) {
				tags = rawArgs[i+1]
				i++
			} else {
				buildArgs = append(buildArgs, rawArgs[i])
			}
		}
		if len(buildArgs) < 1 {
			buildArgs = []string{"."}
		}
		if outputBinary == "" {
			if target == "wasm" || target == "wasm-edge" {
				outputBinary = "service.wasm"
			} else {
				outputBinary = "service.exe"
			}
		}
		buildServ(buildArgs[0], outputBinary, target, goos, goarch, tags)

	case "run":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		watchFlag := runCmd.Bool("watch", false, "Watch files and restart")
		hotFlag := runCmd.Bool("hot", false, "Watch files and hot-reload without restart (zero downtime)")
		profileFlag := runCmd.Bool("profile", false, "Enable CPU and memory profiling")
		envFlag := runCmd.String("env", "", "Environment profile (e.g., staging, production)")
		if err := runCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		args := runCmd.Args()
		if len(args) < 1 {
			args = []string{"."}
		}

		if *hotFlag {
			runServHot(args[0], *envFlag)
		} else if *watchFlag {
			runServWatch(args[0], *envFlag)
		} else {
			runServ(args[0], args[1:], *profileFlag, *envFlag)
		}

	case "dockerize":
		dockerCmd := flag.NewFlagSet("dockerize", flag.ExitOnError)
		if err := dockerCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		args := dockerCmd.Args()
		if len(args) < 1 {
			args = []string{"."}
		}
		dockerizeServ(args[0])

	case "deploy":
		deployCmd := flag.NewFlagSet("deploy", flag.ExitOnError)
		targetFlag := deployCmd.String("target", "", "Deploy target: fly, railway, render, k8s, docker")
		if err := deployCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		if *targetFlag == "" {
			fmt.Println("Usage: pranor deploy --target <fly|railway|render|k8s|docker> [file.pnr]")
			os.Exit(1)
		}
		args := deployCmd.Args()
		if len(args) < 1 {
			args = []string{"."}
		}
		deployServ(args[0], *targetFlag)

	case "test":
		testCmd := flag.NewFlagSet("test", flag.ExitOnError)
		coverFlag := testCmd.Bool("cover", false, "Report test coverage")
		filterFlag := testCmd.String("filter", "", "Filter tests by name")
		integrationFlag := testCmd.Bool("integration", false, "Run with live infrastructure services")
		watchFlag := testCmd.Bool("watch", false, "Watch for changes and re-run tests")
		if err := testCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		args := testCmd.Args()
		if len(args) < 1 {
			args = []string{"."}
		}
		if *watchFlag {
			runTestsWatch(args[0], *coverFlag, *filterFlag, *integrationFlag)
		} else if *integrationFlag {
			if !runIntegrationTests(args[0], *coverFlag, *filterFlag) {
				os.Exit(1)
			}
		} else {
			if !runTests(args[0], *coverFlag, *filterFlag) {
				os.Exit(1)
			}
		}

	case "lint":
		lintCmd := flag.NewFlagSet("lint", flag.ExitOnError)
		if err := lintCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		args := lintCmd.Args()
		if len(args) < 1 {
			args = []string{"."}
		}
		runLint(args[0])

	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		if err := addCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		args := addCmd.Args()
		if len(args) < 1 {
			fmt.Println("Usage: pranor add <go-package-path>")
			fmt.Println("Example: pranor add github.com/google/uuid")
			fmt.Println("         pranor add encoding/base64")
			os.Exit(1)
		}
		addPackage(args[0])

	case "packages":
		listPackages()

	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pranor remove <package-name>")
			os.Exit(1)
		}
		removePackage(os.Args[2])

	case "fmt":
		fmtCmd := flag.NewFlagSet("fmt", flag.ExitOnError)
		checkOnly := fmtCmd.Bool("check", false, "Check if file is formatted (exit 1 if not)")
		if err := fmtCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		args := fmtCmd.Args()
		if len(args) < 1 {
			fmt.Println("Usage: pranor fmt [--check] <file.pnr>")
			os.Exit(1)
		}
		formatFile(args[0], *checkOnly)

	case "repl":
		replCmd := flag.NewFlagSet("repl", flag.ExitOnError)
		attachFlag := replCmd.String("attach", "", "Connect to a running service context (host:port)")
		if len(os.Args) > 2 {
			replCmd.Parse(os.Args[2:])
		}
		runREPL(*attachFlag)

	case "install":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pranor install <package-name>")
			os.Exit(1)
		}
		installPackage(os.Args[2])

	case "publish":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pranor publish <package-dir>")
			os.Exit(1)
		}
		publishPackage(os.Args[2])

	case "init":
		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		fullStackFlag := initCmd.Bool("full-stack", false, "Generate docker-compose.yml with all Pranor services")
		if err := initCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		if *fullStackFlag {
			runInitFullStack()
		} else {
			initProject()
		}

	case "new":
		newCmd := flag.NewFlagSet("new", flag.ExitOnError)
		templateFlag := newCmd.String("template", "api", "Template to scaffold: api, worker, event-processor, full-stack")
		if err := newCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		args := newCmd.Args()
		if len(args) < 1 {
			fmt.Println("Usage: pranor new <project-name> [--template <api|worker|event-processor|full-stack>]")
			os.Exit(1)
		}
		createNewProject(args[0], *templateFlag)

	case "create":
		createCmd := flag.NewFlagSet("create", flag.ExitOnError)
		fixFlag := createCmd.Bool("fix", false, "Repairs failures automatically using pranor test results")
		if err := createCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		args := createCmd.Args()
		if len(args) < 1 {
			fmt.Println("Usage: pranor create [--fix] \"<prompt describing your service>\"")
			os.Exit(1)
		}
		runAIScaffold(args[0], *fixFlag)

	case "debug":
		targetFile := "."
		if len(os.Args) >= 3 {
			targetFile = os.Args[2]
		}
		debugServ(targetFile)

	case "audit":
		runAudit()

	case "doctor":
		doctorCmd := flag.NewFlagSet("doctor", flag.ExitOnError)
		integrationFlag := doctorCmd.Bool("integration", false, "Run full pipeline docker integration checks")
		if err := doctorCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		runDoctor(*integrationFlag)

	case "upgrade":
		runUpgrade()


	case "cache":
		if len(os.Args) >= 3 && os.Args[2] == "inspect" {
			runCacheInspect()
		} else {
			fmt.Println("Usage: pranor cache inspect [--host <host>]")
		}

	case "cron":
		if len(os.Args) >= 3 && os.Args[2] == "list" {
			runCronList()
		} else {
			fmt.Println("Usage: pranor cron list [--host <host>]")
		}

	case "status":
		runStatus()

	case "changelog":
		runChangelog()

	case "monitor":
		target := "8080"
		if len(os.Args) >= 3 {
			target = os.Args[2]
		}
		runMonitor(target)

	case "docs", "doc":
		runDocs()

	case "generate":
		// DX.21: support --from-openapi alias for 'pranor generate routes'
		for _, arg := range os.Args[2:] {
			if arg == "--from-openapi" {
				// Rewrite args: pranor generate --from-openapi <spec> → pranor generate routes <spec>
				newArgs := []string{os.Args[0], "generate", "routes"}
				for _, a := range os.Args[2:] {
					if a != "--from-openapi" {
						newArgs = append(newArgs, a)
					}
				}
				os.Args = newArgs
				break
			}
		}
		runGenerate()

	case "bench":
		runBench()

	case "observability", "obs":
		runObservabilityCmd()

	case "tunnel":
		runTunnelCmd()

	case "config":
		runConfigCmd()

	case "trace":
		runTraceCmd()

	case "playground":
		runPlayground()

	case "queue":
		subcmd := ""
		if len(os.Args) >= 3 {
			subcmd = os.Args[2]
		}
		switch subcmd {
		case "tail":
			runQueueTail()
		case "list":
			runQueueList()
		case "dlq":
			runQueueDLQ()
		default:
			fmt.Println("Usage:")
			fmt.Println("  pranor queue tail <topic> [--host <url>] [--limit <n>]")
			fmt.Println("  pranor queue list [--host <url>]")
			fmt.Println("  pranor queue dlq inspect <topic> [--host <url>] [--replay]")
		}

	case "mesh":
		subcmd := ""
		if len(os.Args) >= 3 {
			subcmd = os.Args[2]
		}
		switch subcmd {
		case "inspect":
			runMeshInspect()
		case "routes":
			runMeshRoutes()
		default:
			fmt.Println("Usage:")
			fmt.Println("  pranor mesh inspect [--host <url>] [--service <name>]")
			fmt.Println("  pranor mesh routes [--host <url>]")
		}

	case "dev":
		runDevCmd()

	case "migrate":
		migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
		dbFlag := migrateCmd.String("db", "", "Database connection string (e.g. sqlite://mydb.db). Falls back to $DATABASE_URL")
		rollbackFlag := migrateCmd.Bool("rollback", false, "Roll back structural schema changes (e.g., dropping columns or tables)")
		dryRunFlag := migrateCmd.Bool("dry-run", false, "Preview structural schema changes with a colored diff without executing them")
		statusFlag := migrateCmd.Bool("status", false, "Show current applied database migrations status")
		if err := migrateCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		args := migrateCmd.Args()
		target := "."
		if len(args) >= 1 {
			target = args[0]
		}
		if *statusFlag {
			runMigrateStatus(target, *dbFlag)
		} else {
			runMigrate(target, *dbFlag, *rollbackFlag, *dryRunFlag)
		}

	case "lsp-action":
		runLspActionCmd(os.Args[2:])

	case "plugin":
		runPluginCmd()

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Serv: A Programming Language for Background Services")
	fmt.Println("Usage:")
	fmt.Println("  pranor init [name]                           Create a new Serv project")
	fmt.Println("  pranor new <name> [--template <template>]    Create a new Serv project from a template (api, worker, event-processor, full-stack)")
	fmt.Println("  pranor create \"<prompt>\"                     AI-scaffold a new Serv file from a natural language description")
	fmt.Println("  pranor docs generate <file.pnr> [-o <out>]   Autogenerate OpenAPI 3.1 specifications from routes")
	fmt.Println("  pranor generate client <file.pnr> [--lang <lang>] [-o <out>] Autogenerate client SDKs (typescript/python/go) from routes")

	fmt.Println("  pranor build <file.pnr> [--target <target>] [-o <output>] Compile Serv code to target (native/wasm)")
	fmt.Println("  pranor run <file.pnr> [--watch]              Compile and run Serv code immediately (with optional hot reload)")
	fmt.Println("  pranor test [--cover] [--integration] <file.pnr> Run tests (--integration starts live infra)")
	fmt.Println("  pranor lint <file.pnr>                       Validate syntax and check for errors")
	fmt.Println("  pranor fmt <file.pnr>                        Format a Serv file")
	fmt.Println("  pranor repl                                  Interactive shell for quick experiments")
	fmt.Println("  pranor add <go-package>                      Generate .pnr.d declaration for a Go package")
	fmt.Println("  pranor install <package-name>                Install a third-party Serv module")
	fmt.Println("  pranor publish <package-dir>                 Publish a Serv module to the registry")
	fmt.Println("  pranor plugin pull <plugin-name>             Pull pre-compiled WASM plugin binary from vyuvaraj/pranor-wasm-plugins")
	fmt.Println("  pranor plugin list                           List community WASM plugins and local installed plugins")
	fmt.Println("  pranor debug <file.pnr>                       Debug a Serv file (requires dlv: go install github.com/go-delve/delve/cmd/dlv@latest)")
	fmt.Println("  pranor dockerize <file.pnr>                  Generate a Dockerfile for the Serv service")
	fmt.Println("  pranor deploy --target <target> [file.pnr]   Generate deploy config (fly, railway, render, k8s, docker)")
	fmt.Println("  pranor monitor [port-or-url]                 Terminal htop-style live dashboard for a running service")
	fmt.Println("  pranor tunnel <port> [options]               Expose a local service via the Pranor Tunnel relay server")
	fmt.Println("  pranor dev [file.pnr] [--dashboard]         Start full dev environment (with optional terminal TUI)")
	fmt.Println("  pranor config init                          Create starter .pranor/config.yaml")
	fmt.Println("  pranor config propagate [--dry-run]         Push configuration settings to active services")
	fmt.Println("  pranor audit                                 Audit Go/Serv dependencies for vulnerabilities")
	fmt.Println("  pranor doctor                                Run compatibility and health checks on all Pranor services")
	fmt.Println("  pranor status                                Print live health, uptime, and latency stats for all services")
	fmt.Println("  pranor migrate [file.pnr] [--db <conn>] [--dry-run] Apply declarative `table` schema migrations to the database")
	fmt.Println("  pranor queue tail <topic> [--host <url>] [--limit <n>]   Tail recent messages on a Pranor Pulse topic")
	fmt.Println("  pranor queue list [--host <url>]                         List all Pranor Pulse topics and consumer counts")
	fmt.Println("  pranor mesh inspect [--host <url>] [--service <name>]   Inspect Pranor Mesh registry and instance list")
	fmt.Println("  pranor mesh routes [--host <url>]                        Show active routing rules and circuit-breaker state")
	fmt.Println("  pranor init --full-stack                                 Generate docker-compose.yml with all Pranor services")
	fmt.Println("  pranor bench <file.pnr> [--rps <n>] [--duration <s>]       Run built-in load test against declared routes")
	fmt.Println("  pranor generate --from-openapi <spec.yaml> [-o <out.pnr>]  Generate .pnr route stubs from OpenAPI spec")
	fmt.Println("  pranor observability init                                   Create .pranor/observability.yaml template")
	fmt.Println("  pranor observability apply [--dry-run]                     Provision alert rules, SLOs, and dashboards")
	fmt.Println("  pranor playground [--port <p>]                             Start the hosted browser-based editor")
	fmt.Println("  pranor lsp-action --file <file> --line <line> [--type <type>] Resolve LSP code action recommendation")
}
