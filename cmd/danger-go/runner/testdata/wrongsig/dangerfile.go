package main

import danger "github.com/danger/golang"

// Run has the wrong signature: it is missing the DSL argument.
func Run(d *danger.T) {}
