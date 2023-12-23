package main

import (
	"os"

	"github.com/frontierdigital/utils/output"
	"github.com/frontierdigital/vertag/pkg/cmd/vertag"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	command := vertag.NewRootCmd(version, commit, date)
	if err := command.Execute(); err != nil {
		output.PrintlnError(err)
		os.Exit(1)
	}
}
