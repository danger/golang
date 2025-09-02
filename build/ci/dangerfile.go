package main

import (
	"fmt"

	danger "github.com/danger/golang"
)

// Run is invoked by danger-go
func Run(d *danger.T, pr danger.DSL) {
	d.Message(fmt.Sprintf("%d new files added!", len(pr.Git.GetCreatedFiles())), "", 0)
	d.Message(fmt.Sprintf("%d files modified!", len(pr.Git.GetModifiedFiles())), "", 0)
}
