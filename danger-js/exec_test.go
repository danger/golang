package dangerJs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

// chdir switches the working directory for the duration of the test, since
// DiffForFile shells out to git in the current directory.
func chdir(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(wd) })
}

func TestDiffForFile(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "git", "init", "-q")
	write := func(content string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte(content), 0o644))
	}
	write("one\ntwo\nthree\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "-dash.txt"), []byte("a\n"), 0o644))
	run(t, dir, "git", "add", "file.txt", "--", "-dash.txt")
	run(t, dir, "git", "commit", "-q", "-m", "first")
	write("one\nTWO\nthree\nfour\n")
	run(t, dir, "git", "commit", "-q", "-am", "second")
	write("zero\none\nTWO\nthree\nfour\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "-dash.txt"), []byte("a\nb\n"), 0o644))
	run(t, dir, "git", "commit", "-q", "-am", "third")
	chdir(t, dir)

	g := gitImpl{}

	t.Run("defaults to HEAD^..HEAD", func(t *testing.T) {
		diff, err := g.DiffForFile("file.txt")
		require.NoError(t, err)
		require.Equal(t, FileDiff{AddedLines: []DiffLine{{Content: "zero", Line: 1}}}, diff)
	})

	t.Run("explicit refs", func(t *testing.T) {
		diff, err := g.DiffForFileWithRefs("file.txt", "HEAD~2", "HEAD~1")
		require.NoError(t, err)
		require.Equal(t, FileDiff{
			AddedLines:   []DiffLine{{Content: "TWO", Line: 2}, {Content: "four", Line: 4}},
			RemovedLines: []DiffLine{{Content: "two", Line: 2}},
		}, diff)
	})

	t.Run("path starting with a dash is passed after --", func(t *testing.T) {
		diff, err := g.DiffForFile("-dash.txt")
		require.NoError(t, err)
		require.Equal(t, FileDiff{AddedLines: []DiffLine{{Content: "b", Line: 2}}}, diff)
	})

	t.Run("unchanged file", func(t *testing.T) {
		diff, err := g.DiffForFileWithRefs("file.txt", "HEAD", "HEAD")
		require.NoError(t, err)
		require.Equal(t, FileDiff{}, diff)
	})

	t.Run("unknown ref", func(t *testing.T) {
		_, err := g.DiffForFileWithRefs("file.txt", "no-such-ref", "HEAD")
		require.Error(t, err)
	})

	outFile := filepath.Join(t.TempDir(), "out")
	rejected := []struct {
		name, path, base, head, errPrefix string
	}{
		{name: "path injection", path: "file.txt; rm -rf /", base: "HEAD^", head: "HEAD", errPrefix: "invalid file path"},
		{name: "base ref injection", path: "file.txt", base: "HEAD;id", head: "HEAD", errPrefix: "invalid base ref"},
		{name: "head ref injection", path: "file.txt", base: "HEAD^", head: "--output=" + outFile, errPrefix: "invalid head ref"},
	}
	for _, c := range rejected {
		t.Run(c.name, func(t *testing.T) {
			_, err := g.DiffForFileWithRefs(c.path, c.base, c.head)
			require.Error(t, err)
			require.True(t, strings.HasPrefix(err.Error(), c.errPrefix), err.Error())
			require.NoFileExists(t, outFile, "git must not have been run with an injected option")
		})
	}
}

// fakeDanger puts a `danger` script on PATH that records the arguments it
// was called with, so Process can be tested without danger-js installed.
func fakeDanger(t *testing.T) (binDir, argsFile string) {
	t.Helper()
	binDir = t.TempDir()
	argsFile = filepath.Join(binDir, "args.txt")
	script := "#!/bin/sh\nfor a in \"$@\"; do echo \"$a\" >> " + argsFile + "; done\n"
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "danger"), []byte(script), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "danger-go"), []byte("#!/bin/sh\n"), 0o755))
	return binDir, argsFile
}

func TestProcess(t *testing.T) {
	cc := []struct {
		name    string
		command string
		args    []string
	}{
		{name: "ci", command: "ci"},
		{name: "local", command: "local", args: []string{"--base", "main"}},
		{name: "pr", command: "pr", args: []string{"https://github.com/danger/golang/pull/46"}},
	}
	for _, c := range cc {
		t.Run(c.name, func(t *testing.T) {
			binDir, argsFile := fakeDanger(t)
			t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

			require.NoError(t, Process(c.command, c.args))

			got, err := os.ReadFile(argsFile)
			require.NoError(t, err)
			exp := append([]string{c.command, "--process", filepath.Join(binDir, "danger-go"), "--passURLForDSL"}, c.args...)
			require.Equal(t, strings.Join(exp, "\n")+"\n", string(got))
		})
	}
}

func TestProcessMissingBinaries(t *testing.T) {
	t.Run("no danger", func(t *testing.T) {
		// Keep /usr/bin:/bin so `which` itself runs and reports danger missing.
		t.Setenv("PATH", t.TempDir()+string(os.PathListSeparator)+"/usr/bin:/bin")
		err := Process("ci", nil)
		require.ErrorContains(t, err, "could not find `danger` binary")
	})

	t.Run("no danger-go", func(t *testing.T) {
		binDir, _ := fakeDanger(t)
		require.NoError(t, os.Remove(filepath.Join(binDir, "danger-go")))
		t.Setenv("PATH", binDir+string(os.PathListSeparator)+"/usr/bin:/bin")
		err := Process("ci", nil)
		require.ErrorContains(t, err, "could not find `danger-go` binary")
	})
}

func TestProcessPropagatesDangerFailure(t *testing.T) {
	binDir, _ := fakeDanger(t)
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "danger"), []byte("#!/bin/sh\nexit 1\n"), 0o755))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	require.Error(t, Process("ci", nil), "a failing danger run must fail danger-go so CI goes red")
}

// fakeDangerPR writes a `danger` script that behaves like `danger pr --json`:
// progress on stderr, the DSL on stdout, and its arguments recorded.
func fakeDangerPR(t *testing.T, body string) (bin, argsFile string) {
	t.Helper()
	dir := t.TempDir()
	bin = filepath.Join(dir, "danger")
	argsFile = filepath.Join(dir, "args.txt")
	script := "#!/bin/sh\nfor a in \"$@\"; do echo \"$a\" >> " + argsFile + "; done\n" + body
	require.NoError(t, os.WriteFile(bin, []byte(script), 0o755))
	return bin, argsFile
}

func TestGetPR(t *testing.T) {
	fixture, err := filepath.Abs("testdata/github_pr_dsl.json")
	require.NoError(t, err)
	bin, argsFile := fakeDangerPR(t, "echo 'Starting Danger PR on danger/golang#46' >&2\ncat "+fixture+"\n")

	dsl, err := GetPR("https://github.com/danger/golang/pull/46", bin)
	require.NoError(t, err)
	require.Equal(t, 46, dsl.GitHub.PR().Number)
	require.Equal(t, []FilePath{".github/workflows/test.yml"}, dsl.Git.ModifiedFiles())

	args, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	require.Equal(t, "pr\nhttps://github.com/danger/golang/pull/46\n--json\n--dangerfile\ndangerfile.go\n", string(args))
}

func TestGetPRFindsDangerOnPath(t *testing.T) {
	fixture, err := filepath.Abs("testdata/github_pr_dsl.json")
	require.NoError(t, err)
	bin, _ := fakeDangerPR(t, "cat "+fixture+"\n")
	t.Setenv("PATH", filepath.Dir(bin)+string(os.PathListSeparator)+os.Getenv("PATH"))

	dsl, err := GetPR("https://github.com/danger/golang/pull/46", "")
	require.NoError(t, err)
	require.Equal(t, 46, dsl.GitHub.PR().Number)
}

func TestGetPRErrors(t *testing.T) {
	t.Run("danger fails", func(t *testing.T) {
		bin, _ := fakeDangerPR(t, "echo 'Bad credentials' >&2\nexit 1\n")
		_, err := GetPR("https://github.com/danger/golang/pull/46", bin)
		require.ErrorContains(t, err, "could not download DSL JSON with danger-js")
		require.ErrorContains(t, err, "Bad credentials")
	})

	t.Run("output is not JSON", func(t *testing.T) {
		bin, _ := fakeDangerPR(t, "echo 'not json'\n")
		_, err := GetPR("https://github.com/danger/golang/pull/46", bin)
		require.ErrorContains(t, err, "decoding DSL JSON from danger-js")
	})
}
