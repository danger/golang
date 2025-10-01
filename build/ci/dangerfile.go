package main

import (
	"fmt"

	danger "github.com/danger/golang"
)

// Run is invoked by danger-go
func Run(d *danger.T, pr danger.DSL) {
	d.Message(fmt.Sprintf("%d new files added!", len(pr.Git.CreatedFiles())), "", 0)
	d.Message(fmt.Sprintf("%d files modified!", len(pr.Git.ModifiedFiles())), "", 0)
}
