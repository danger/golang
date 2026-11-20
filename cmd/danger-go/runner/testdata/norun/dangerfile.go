package main

import danger "github.com/danger/golang"

// Lint is not Run, so the runner cannot find the entrypoint.
func Lint(d *danger.T, pr danger.DSL) {}
