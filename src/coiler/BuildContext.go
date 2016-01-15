package coiler

import (
	"strings"
	"fmt"
)

const (
	NAMESPACE_SEPARATOR = "_ZC_"
)

/*
	A BuildContext is used to maintain knowledge about the current state of a build.
*/
type BuildContext struct {

	// A graph that represents all dependencies (and their interrelations) used in this build run.
	dependencies *DependencyGraph

	// A list of non-combined import modules which need to be included in the final combined output
	externalDependencies []string

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

/*
	Takes a fully-qualified symbol name, and creates a namespaced version suitable for use in combined files.
*/
func (this *BuildContext) AddSymbol(symbol string) {

	this.symbols[symbol] = strings.Replace(symbol, ".", NAMESPACE_SEPARATOR, -1)
}

func (this *BuildContext) AddImportedFile(module string) {

	if(!this.IsFileImported(module)) {
		this.importedFiles = append(this.importedFiles, module)
	}
}

func (this *BuildContext) IsFileImported(module string) bool {

	for _, file := range this.importedFiles {
		if(module == file) {
			return true
		}
	}
	return false
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

func (this *BuildContext) AddExternalDependency(module string) {
	this.externalDependencies = append(this.externalDependencies, module)
}

func (this *BuildContext) String() string {

	return fmt.Sprintf("%v", this.externalDependencies)
}
