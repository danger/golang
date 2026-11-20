package dangerJs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	dangerJsBinary = "danger"
	dangerGoBinary = "danger-go"
)

func findBinary(name string) (string, error) {
	cmd := exec.Command("which", name)
	dangerBin, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not find `%s` binary: %w", name, err)
	}
	return strings.TrimSpace(string(dangerBin)), nil
}

// GetPR fetches the DSL for an existing PR with `danger pr <url> --json`.
// If dangerBin is empty, `danger` is looked up on PATH.
func GetPR(url string, dangerBin string) (DSL, error) {
	var err error
	if dangerBin == "" {
		dangerBin, err = findBinary(dangerJsBinary)
		if err != nil {
			return DSL{}, err
		}
	}

	// Without --dangerfile, danger-js refuses to run unless a dangerfile.js or
	// dangerfile.ts exists, even though --json never reads it.
	cmd := exec.Command(dangerBin, "pr", url, "--json", "--dangerfile", "dangerfile.go")
	// danger-js writes progress to stderr, so only stdout is JSON.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	prJSON, err := cmd.Output()
	if err != nil {
		return DSL{}, fmt.Errorf("could not download DSL JSON with danger-js: %w: %s", err, stderr.String())
	}

	var prData struct {
		Danger DSLData `json:"danger"`
	}
	if err = json.Unmarshal(prJSON, &prData); err != nil {
		return DSL{}, fmt.Errorf("decoding DSL JSON from danger-js: %w", err)
	}
	return prData.Danger.ToInterface(), nil
}

func Process(command string, args []string) error {
	dangerBin, err := findBinary(dangerJsBinary)
	if err != nil {
		return err
	}
	dangerGoBin, err := findBinary(dangerGoBinary)
	if err != nil {
		return err
	}
	// The `danger` (javascript) command will call the process specified,
	// i.e. `danger-go`, with the first argument of `runner` followed by the
	// arguments it received.
	cmdArgs := append([]string{command, "--process", dangerGoBin, "--passURLForDSL"}, args...)
	cmd := exec.Command(dangerBin, cmdArgs...)
	fmt.Printf("Running: %s\n", cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
