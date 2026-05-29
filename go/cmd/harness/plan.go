package main

import (
	"fmt"
	"os"
)

// runPlanCLI is the entry point for the "harness plan-cli" subcommand.
// Full implementation in this file; called from main.go.
func runPlanCLI(args []string) {
	if len(args) == 0 {
		planUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "list", "ls":
		runPlanList(args[1:])
	case "get":
		runPlanGet(args[1:])
	case "add-phase":
		runPlanAddPhase(args[1:])
	case "add-task":
		runPlanAddTask(args[1:])
	case "update":
		runPlanUpdate(args[1:])
	case "archive":
		runPlanArchive(args[1:])
	case "comment":
		runPlanComment(args[1:])
	case "migrate":
		runPlanMigrate(args[1:])
	case "serve":
		runPlanServe(args[1:])
	case "help", "--help", "-h":
		planUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown plan-cli subcommand: %s\n", args[0])
		planUsage()
		os.Exit(1)
	}
}

func planUsage() {
	fmt.Fprintln(os.Stderr, "Usage: harness plan-cli <subcommand>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Subcommands:")
	fmt.Fprintln(os.Stderr, "  list (ls)              List phases/tasks (default: active phases)")
	fmt.Fprintln(os.Stderr, "    --status  active|archived|all|cc:TODO|cc:WIP|cc:done|pm:confirmed|blocked")
	fmt.Fprintln(os.Stderr, "    --urgency low|medium|high")
	fmt.Fprintln(os.Stderr, "    --importance low|medium|high")
	fmt.Fprintln(os.Stderr, "    --marker  feature:security|ralph|...")
	fmt.Fprintln(os.Stderr, "    --phase   <id>")
	fmt.Fprintln(os.Stderr, "    --pretty  human-readable table output")
	fmt.Fprintln(os.Stderr, "  get <phase-id|task-id>  Print a single phase or task as JSON")
	fmt.Fprintln(os.Stderr, "  add-phase               Add a new phase")
	fmt.Fprintln(os.Stderr, "    --title  \"string\"  (required)")
	fmt.Fprintln(os.Stderr, "    --goal   \"string\"  (required)")
	fmt.Fprintln(os.Stderr, "    --urgency low|medium|high  (default: medium)")
	fmt.Fprintln(os.Stderr, "    --importance low|medium|high  (default: medium)")
	fmt.Fprintln(os.Stderr, "  add-task <phase-id>     Add a task to a phase")
	fmt.Fprintln(os.Stderr, "    --name        \"string\"  (required)")
	fmt.Fprintln(os.Stderr, "    --description \"string\"")
	fmt.Fprintln(os.Stderr, "    --dod         \"string\"  (required)")
	fmt.Fprintln(os.Stderr, "    --depends     \"1.1,1.2\"")
	fmt.Fprintln(os.Stderr, "    --urgency     low|medium|high")
	fmt.Fprintln(os.Stderr, "    --importance  low|medium|high")
	fmt.Fprintln(os.Stderr, "    --marker      feature:security,skip:tdd,...")
	fmt.Fprintln(os.Stderr, "  update <task-id>        Update task fields")
	fmt.Fprintln(os.Stderr, "    --status   cc:TODO|cc:WIP|cc:done|pm:confirmed|pm:requested|blocked")
	fmt.Fprintln(os.Stderr, "    --hash     \"abc1234\"  (required when status=cc:done)")
	fmt.Fprintln(os.Stderr, "    --reason   \"string\"  (required when status=blocked)")
	fmt.Fprintln(os.Stderr, "    --urgency  low|medium|high")
	fmt.Fprintln(os.Stderr, "    --importance low|medium|high")
	fmt.Fprintln(os.Stderr, "  archive <phase-id>      Set phase status to archived")
	fmt.Fprintln(os.Stderr, "  comment <phase-id|task-id>  Add a comment")
	fmt.Fprintln(os.Stderr, "    --text        \"string\"   (required)")
	fmt.Fprintln(os.Stderr, "    --author      human|agent  (default: human)")
	fmt.Fprintln(os.Stderr, "    --author-name \"string\"")
	fmt.Fprintln(os.Stderr, "  migrate                 Migrate Plans.md → plans.json")
	fmt.Fprintln(os.Stderr, "    --from      ./Plans.md  (default)")
	fmt.Fprintln(os.Stderr, "    --dry-run   print JSON, do not write")
	fmt.Fprintln(os.Stderr, "  serve                   Start local web UI")
	fmt.Fprintln(os.Stderr, "    --port  8888  (default)")
	fmt.Fprintln(os.Stderr, "    --open  open browser on start")
}

// planPath resolves plans.json path relative to cwd.
func planPath() (string, error) {
	root, err := resolveProjectRoot(nil)
	if err != nil {
		return "", err
	}
	return DefaultPlansPath(root), nil
}

// planLoad loads plans.json or returns an empty Plans if not found.
func planLoad() (*Plans, string, error) {
	path, err := planPath()
	if err != nil {
		return nil, "", err
	}
	p, err := LoadPlans(path)
	return p, path, err
}
