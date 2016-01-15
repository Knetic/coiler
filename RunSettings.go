package main

import (
	"flag"
)

type RunSettings struct {

	CombineMode string
	EntryPointPath string
	OutputPath string
}

func ParseRunSettings() RunSettings {

	var ret RunSettings

	flag.StringVar(&ret.CombineMode, "m", "user", "Mode for parsing. 'user' will combine only user and third-party modules together, but 'all' will also include system libraries.")
	flag.StringVar(&ret.OutputPath, "o", "./a.pyc", "Path to output a final '*.pyc' file")
	flag.StringVar(&ret.EntryPointPath, "i", "", "Path to the input entry point")
	flag.Parse()

	return ret
}
