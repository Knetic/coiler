package coiler

/*
	A BuildContext is used to maintain knowledge about the current state of a build.
*/
type BuildContext struct {

	// Contains a mapping of fully-qualified function names and variable names
	// and the translated version of each.
	symbols map[string]string

	// represents every file that has been functionally included (if not necessarily combined into)
	// the output file for this build run.
	importedFiles []string
}
