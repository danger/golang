package dangerJs

import (
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

func GetPR(url string, dangerBin string) (DSL, error) {
	var err error
	if dangerBin == "" {
		dangerBin, err = findBinary(dangerJsBinary)
		if err != nil {
			return DSL{}, err
		}
	}

	cmd := exec.Command(dangerBin, "pr", url, "--json")
	prJSON, err := cmd.CombinedOutput()
	if err != nil {
		return DSL{}, fmt.Errorf("could not download DSL JSON with danger-js: %w", err)
	}

	var prData DSLData
	if err = json.Unmarshal(prJSON, &prData); err != nil {
		return DSL{}, err
	}
	return prData.ToInterface(), nil
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
