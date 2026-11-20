package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const cliPkg = "github.com/danger/golang/cmd/danger-go"

func execute(args ...string) (string, error) {
	cmdArgs := append([]string{"run", cliPkg}, args...)
	cmd := exec.Command("go", cmdArgs...)
	res, err := cmd.CombinedOutput()
	return string(res), err
}

func TestShowsUsage(t *testing.T) {
	cc := []struct {
		name string
		args []string
	}{
		{name: "no args", args: nil},
		{name: "-h arg", args: []string{"-h"}},
		{name: "--help arg", args: []string{"--help"}},
	}

	for _, c := range cc {
		t.Run(c.name, func(t *testing.T) {
			res, err := execute(c.args...)
			require.Nil(t, err)
			require.Equal(t, usage, res)
		})
	}
}

func TestShowsVersion(t *testing.T) {
	res, err := execute("version")
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("danger-go %s\n", version), res)
}

func TestInvalidSubCommand(t *testing.T) {
	res, err := execute("bogus")
	require.Error(t, err)
	require.Contains(t, res, "invalid sub-command `bogus`")
	require.Contains(t, res, usage)
}

func TestHelpFlagAfterCommand(t *testing.T) {
	res, err := execute("ci", "--help")
	require.Nil(t, err)
	require.Equal(t, usage, res)
}

func TestDangerCommandsNeedDangerJS(t *testing.T) {
	for _, command := range []string{"ci", "local", "pr"} {
		t.Run(command, func(t *testing.T) {
			bin := buildCLI(t)
			cmd := exec.Command(bin, command)
			cmd.Env = append(os.Environ(), "PATH="+t.TempDir())
			res, err := cmd.CombinedOutput()
			require.Error(t, err)
			require.Contains(t, string(res), "could not find `danger` binary")
		})
	}
}

// buildCLI builds danger-go once per test so it can run with a modified PATH,
// which `go run` itself needs.
func buildCLI(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "danger-go")
	out, err := exec.Command("go", "build", "-o", bin, cliPkg).CombinedOutput()
	require.NoError(t, err, string(out))
	return bin
}
