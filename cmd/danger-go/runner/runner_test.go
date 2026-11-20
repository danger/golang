package runner_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	danger "github.com/danger/golang"
)

// These tests drive the real `danger-go runner` binary the way danger-js
// does: a danger://dsl/<path> URL on stdin, results JSON on stdout. The
// dangerfile is built as a Go plugin, which must match the host binary's
// toolchain and flags, so the runner runs as a subprocess rather than in the
// test binary.

var dangerGoBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "danger-go-runner-test-")
	if err != nil {
		panic(err)
	}
	dangerGoBin = filepath.Join(dir, "danger-go")
	build := exec.Command("go", "build", "-o", dangerGoBin, "github.com/danger/golang/cmd/danger-go")
	if out, err := build.CombinedOutput(); err != nil {
		panic("building danger-go: " + err.Error() + "\n" + string(out))
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// fixtureDSL is the real danger-js payload shared with the danger-js package
// contract tests.
func fixtureDSL(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs("../../../danger-js/testdata/github_pr_dsl.json")
	require.NoError(t, err)
	return p
}

func fixtureURL(t *testing.T) string {
	return "danger://dsl/" + fixtureDSL(t)
}

type runResult struct {
	stdout, stderr string
	err            error
}

func runRunner(t *testing.T, dangerfileDir, stdin string) runResult {
	t.Helper()
	cmd := exec.Command(dangerGoBin, "runner")
	cmd.Dir = filepath.Join("testdata", dangerfileDir)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return runResult{stdout: stdout.String(), stderr: stderr.String(), err: err}
}

// results parses the JSON danger-js reads back: the runner prints build
// progress first, and the results JSON last with no trailing newline.
func results(t *testing.T, stdout string) danger.Results {
	t.Helper()
	lines := strings.Split(stdout, "\n")
	var r danger.Results
	require.NoError(t, json.Unmarshal([]byte(lines[len(lines)-1]), &r), stdout)
	return r
}

func TestRunnerRunsDangerfile(t *testing.T) {
	// danger-js writes the URL to stdin without a trailing newline.
	res := runRunner(t, "full", fixtureURL(t))
	require.NoError(t, res.err, res.stderr)

	require.Equal(t, danger.Results{
		Messages: []danger.Violation{
			{Message: "modified=.github/workflows/test.yml"},
			{Message: "commits=1"},
			{Message: "pr=46 renovate: Update dependency node to v26"},
			{Message: "thisPR=danger/golang#46"},
			{Message: "verbose=false"},
		},
		Warnings:  []danger.Violation{{Message: "a warning", File: "api.go", Line: 12}},
		Fails:     []danger.Violation{{Message: "a failure", File: "types.go"}},
		Markdowns: []danger.Violation{{Message: "## Some markdown"}},
	}, results(t, res.stdout))
}

func TestRunnerAcceptsSurroundingWhitespace(t *testing.T) {
	res := runRunner(t, "full", "  danger://dsl/"+fixtureDSL(t)+"\n")
	require.NoError(t, res.err, res.stderr)
	require.Len(t, results(t, res.stdout).Messages, 5)
}

// TestRunnerVerboseFlag covers `danger-go ci --verbose`: danger-js sends
// cliArgs.verbose as a JSON boolean.
func TestRunnerVerboseFlag(t *testing.T) {
	b, err := os.ReadFile(fixtureDSL(t))
	require.NoError(t, err)
	var dsl map[string]any
	require.NoError(t, json.Unmarshal(b, &dsl))
	dsl["danger"].(map[string]any)["settings"].(map[string]any)["cliArgs"] = map[string]any{
		"verbose": true, "textOnly": true,
	}
	b, err = json.Marshal(dsl)
	require.NoError(t, err)
	p := filepath.Join(t.TempDir(), "dsl.json")
	require.NoError(t, os.WriteFile(p, b, 0o600))

	res := runRunner(t, "full", "danger://dsl/"+p)
	require.NoError(t, res.err, res.stderr)
	require.Contains(t, results(t, res.stdout).Messages, danger.Violation{Message: "verbose=true"})
}

func TestRunnerFailures(t *testing.T) {
	cc := []struct {
		name       string
		dangerfile string
		stdin      func(t *testing.T) string
		errMsg     string
	}{
		{
			name:       "stdin is not a DSL URL",
			dangerfile: "full",
			stdin:      func(t *testing.T) string { return `{"danger":{}}` },
			errMsg:     "did not receive a DSL URL",
		},
		{
			name:       "DSL file missing",
			dangerfile: "full",
			stdin:      func(t *testing.T) string { return "danger://dsl/" + filepath.Join(t.TempDir(), "missing.json") },
			errMsg:     "failed to read JSON file",
		},
		{
			name:       "DSL file is not JSON",
			dangerfile: "full",
			stdin: func(t *testing.T) string {
				p := filepath.Join(t.TempDir(), "dsl.json")
				require.NoError(t, os.WriteFile(p, []byte("not json"), 0o600))
				return "danger://dsl/" + p
			},
			errMsg: "failed to unmarshal DSL JSON",
		},
		{
			name:       "no dangerfile.go",
			dangerfile: "nodangerfile",
			stdin:      fixtureURL,
			errMsg:     "`dangerfile.go` does not exist",
		},
		{
			name:       "dangerfile does not compile",
			dangerfile: "compileerr",
			stdin:      fixtureURL,
			errMsg:     "building plugin from dangerfile",
		},
		{
			name:       "dangerfile has no Run",
			dangerfile: "norun",
			stdin:      fixtureURL,
			errMsg:     `symbol Run not found`,
		},
		{
			name:       "Run has the wrong signature",
			dangerfile: "wrongsig",
			stdin:      fixtureURL,
			errMsg:     "failed to cast DSL function",
		},
	}
	for _, c := range cc {
		t.Run(c.name, func(t *testing.T) {
			res := runRunner(t, c.dangerfile, c.stdin(t))
			require.Error(t, res.err, "runner must exit non-zero so danger-js reports the failure")
			require.Contains(t, res.stderr, c.errMsg)
		})
	}
}
