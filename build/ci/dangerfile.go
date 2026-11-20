package main

import (
	"fmt"
	"os"

	danger "github.com/danger/golang"
)

// Run is invoked by danger-go
func Run(d *danger.T, pr danger.DSL) {
	// `danger-go local` has no PR data, so only check on CI.
	if os.Getenv("CI") != "" {
		selfCheck(d, pr)
	}

	d.Message(fmt.Sprintf("%d new files added!", len(pr.Git.CreatedFiles())), "", 0)
	d.Message(fmt.Sprintf("%d files modified!", len(pr.Git.ModifiedFiles())), "", 0)
}

// selfCheck fails the build if the DSL danger-js handed over decoded to
// empty values. Every PR changes at least one file and has a number and a
// commit, so an empty field means danger-js and danger-go no longer agree on
// the DSL shape (for example after a Node or danger-js upgrade).
func selfCheck(d *danger.T, pr danger.DSL) {
	changed := len(pr.Git.CreatedFiles()) + len(pr.Git.ModifiedFiles()) + len(pr.Git.DeletedFiles())
	if changed == 0 {
		d.Fail("danger-go self-check: the DSL lists no changed files", "", 0)
	}
	if len(pr.Git.Commits()) == 0 {
		d.Fail("danger-go self-check: the DSL lists no commits", "", 0)
	}
	if pr.GitHub.PR().Number == 0 || pr.GitHub.ThisPR().Number != pr.GitHub.PR().Number {
		d.Fail("danger-go self-check: the DSL has no GitHub PR number", "", 0)
	}
}
