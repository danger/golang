package main

import (
	"fmt"
	"strings"

	danger "github.com/danger/golang"
)

// Run exercises every result type and reads from each part of the DSL, so the
// runner test can check both directions of the danger-js contract.
func Run(d *danger.T, pr danger.DSL) {
	d.Message(fmt.Sprintf("modified=%s", strings.Join(pr.Git.ModifiedFiles(), ",")), "", 0)
	d.Message(fmt.Sprintf("commits=%d", len(pr.Git.Commits())), "", 0)
	d.Message(fmt.Sprintf("pr=%d %s", pr.GitHub.PR().Number, pr.GitHub.PR().Title), "", 0)
	d.Message(fmt.Sprintf("thisPR=%s/%s#%d", pr.GitHub.ThisPR().Owner, pr.GitHub.ThisPR().Repo, pr.GitHub.ThisPR().Number), "", 0)
	d.Message(fmt.Sprintf("verbose=%t", pr.Settings.CLIArgs().Verbose), "", 0)
	d.Warn("a warning", "api.go", 12)
	d.Fail("a failure", "types.go", 0)
	d.Markdown("## Some markdown", "", 0)
}
