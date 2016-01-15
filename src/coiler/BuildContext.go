package coiler

/*
	A BuildContext is used to maintain knowledge about the current state of a build.
*/
type BuildContext struct {

	// A graph that represents all dependencies (and their interrelations) used in this build run.
	dependencies *DependencyGraph

	// Contains a mapping of fully-qualified function names and variable names
	// and the translated version of each.
	symbols map[string]string

	// represents every file that has been functionally included (if not necessarily combined into)
	// the output file for this build run.
	importedFiles []string

	// the paths used for external module lookups.
	// different modes of operation mutate this.
	lookupPaths []string
}

func NewBuildContext() *BuildContext {

	var ret *BuildContext

	ret = new(BuildContext)
	ret.dependencies = NewDependencyGraph()

	return ret
}

func (this *BuildContext) WriteCombinedOutput(target string) error {
	return nil
}

/*
	Searches this context's lookup paths to find the appropriate file to provide the given [module].
*/
func (this *BuildContext) FindSourcePath(module string) string {

	return ""
}
