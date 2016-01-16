package main

import (
	"coiler"
	"fmt"
	"os"
)

func main() {

	var settings RunSettings
	var err error

	settings = ParseRunSettings()

	err = coiler.Parse(settings.EntryPointPath, settings.OutputPath, settings.CombineMode == "all")
	if(err != nil) {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
		return
	}
}
