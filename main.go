package main

import (
	"coiler"
	"fmt"
	"os"
	"time"
)

func main() {

	var settings RunSettings
	var context *coiler.BuildContext
	var startTime, currentTime time.Time
	var elapsed int64
	var err error

	startTime = time.Now()
	settings = ParseRunSettings()

	context, err = coiler.Parse(settings.EntryPointPath, settings.CombineMode == "all")
	if(err != nil) {
		printError(1, "Unable to parse source files: \n%v\n", err)
		return
	}

	currentTime = time.Now()
	elapsed = (currentTime.Unix() - startTime.Unix())
	fmt.Printf("Took %dms to parse %d files\n", elapsed, context.GetCombinedFileCount())

	startTime = currentTime
	err = coiler.CompileCombinedFile(settings.OutputPath, context)
	if(err != nil) {
		printError(1, "\nUnable to compile combined output: \n%v\n", err)
		return
	}

	currentTime = time.Now()
	elapsed = (currentTime.Unix() - startTime.Unix())
	fmt.Printf("Took %dms to compile\n", elapsed)
}

func printError(status int, format string, args ...interface{}) {

	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(status)
}
