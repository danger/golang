package main

import danger "github.com/danger/golang"

func Run(d *danger.T, pr danger.DSL) {
	d.Message(undefinedVariable, "", 0)
}
