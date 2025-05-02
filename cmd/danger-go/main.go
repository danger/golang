package main

import (
	"fmt"
	"log"
	"os"

	"github.com/danger/golang/cmd/danger-go/runner"
	dangerJs "github.com/danger/golang/danger-js"
)

const version = "v0.3.0"

// main entrypoint of the danger-go command
func main() {
	if len(os.Args) <= 1 || argsContain("-h", "--help") {
		fmt.Print(usage)
		return
	}

	command := os.Args[1]
	switch command {
	case "ci", "local", "pr":
		var rest []string
		if len(os.Args) > 2 {
			rest = os.Args[2:]
		}
		err := dangerJs.Process(command, rest)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	case "runner":
		runner.Run()
	case "version":
		fmt.Printf("danger-go %s\n", version)
	default:
		log.Fatalf("invalid sub-command `%s`\n\n%s", command, usage)
	}
}

// argsContain returns true if any of the provided `args` are in the list passed
// in to the command
func argsContain(args ...string) bool {
	for _, a := range os.Args {
		for _, v := range args {
			if a == v {
				return true
			}
		}
	}
	return false
}

const usage = `Usage: danger-go [options] [command]

Options:
  -h, --help     Output usage information

Commands:
  ci             Runs DSL on CI
  local          Runs danger standalone on a repo, useful for git hooks
  pr             Runs your local Dangerfile against an existing GitHub DSL. Will not post on the DSL
  runner         Runs a dangerfile against a DSL passed in via STDIN [You probably don't need this]
  version        Show the version of the application
`
