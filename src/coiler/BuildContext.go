package coiler

import (
	"strings"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"io"
	"path/filepath"
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
	// keys are module names, values are absolute paths to the source files for them
	lookupFiles map[string]string
}

func NewBuildContext(useSystemPaths bool) *BuildContext {

	var ret *BuildContext

	ret = new(BuildContext)
	ret.symbols = make(map[string]string)
	ret.dependencies = NewDependencyGraph()
	ret.lookupFiles = determineLookupFiles(determineLookupPaths(useSystemPaths))

	return ret
}

/*
	Takes a fully-qualified symbol name, and creates a namespaced version suitable for use in combined files.
	Returns the translated name.
*/
func (this *BuildContext) AddSymbol(symbol string) string {

	var translatedSymbol string

	translatedSymbol = strings.Replace(symbol, ".", NAMESPACE_SEPARATOR, -1)
	this.symbols[symbol] = translatedSymbol
	return translatedSymbol
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

/*
	Takes the current build context and writes a single combined source file to the given [targetPath]
*/
func (this *BuildContext) WriteCombinedOutput(targetPath string) error {

	var fileContexts []*FileContext
	var outFile *os.File
	var line string
	var err error

	outFile, err = os.Create(targetPath)
	if(err != nil) {

		outFile, err = os.Open(targetPath)
		if(err != nil) {
			return err
		}
	}
	defer outFile.Close()

	// write external dependencies first
	for _, dependency := range this.externalDependencies {

		line = fmt.Sprintf("import %v\n", dependency)
		outFile.Write([]byte(line))
	}

	this.dependencies.DiscoverNeighbors()
	fileContexts = this.dependencies.GetOrderedNodes()

	for _, context := range fileContexts {

		fmt.Printf("Writing context: %s\n", context.namespace)
		err = this.writeTranslatedFile(context, outFile)
		if(err != nil) {
			return err
		}
	}
	return nil
}

func (this *BuildContext) writeTranslatedFile(context *FileContext, outFile *os.File) error {

	var sourceFile *os.File
	var err error

	sourceFile, err = os.Open(context.fullPath)
	if(err != nil) {
		return err
	}
	defer sourceFile.Close()

	written, err := io.Copy(outFile, sourceFile)
	if(err != nil) {
		return err
	}

	fmt.Printf("Wrote %d bytes from '%s'\n", written, context.fullPath)
	return nil
}

/*
	Searches this context's lookup paths to find the appropriate file to provide the given [module].
*/
func (this *BuildContext) FindSourcePath(module string) string {

	return this.lookupFiles[module]
}

func (this *BuildContext) AddDependency(context *FileContext) {
	this.dependencies.AddNode(context)
}

func (this *BuildContext) AddExternalDependency(module string) {

	for _, dependency := range this.externalDependencies {
		if(dependency == module) {
			return
		}
	}

	this.externalDependencies = append(this.externalDependencies, module)
}

func (this *BuildContext) String() string {
	return fmt.Sprintf("External: %v\nImported: %v\nSymbols: %v\n", this.externalDependencies, this.importedFiles, this.symbols)
}

/*
	Determines the python lookup paths to use.
*/
func determineLookupPaths(useSystemPaths bool) []string {

	var process *exec.Cmd
	var pathExtractor *regexp.Regexp
	var output []byte
	var paths []string
	var matches [][]string
	var pyPaths string

	process = exec.Command("python", "-c", "import sys; print(sys.path)")
	output, _ = process.Output()
	pyPaths = string(output)

	pathExtractor = regexp.MustCompile("'([^']*)'")
	matches = pathExtractor.FindAllStringSubmatch(pyPaths, -1)

	for _, match := range matches {

		// ignore egg files for right now
		if(!strings.HasSuffix(match[1], ".egg")) {
			paths = append(paths, match[1])
		}

		// if we do not use system paths, trim out any paths that are not descended from PYTHONPATH or local.
	}

	return paths
}

/*
	Given a list of directories, finds all python source files. Does not recurse.
	Returns a map of module names to absolute paths.
*/
func determineLookupFiles(paths []string) map[string]string {

	var ret map[string]string
	var sourceFiles []string
	var directoryName string
	var fullPath, module string
	var err error

	ret = make(map[string]string)

	for _, path := range paths {

		directoryName = filepath.Join(path, "*.py")
		sourceFiles, err = filepath.Glob(directoryName)
		if(err != nil) {
			fmt.Printf("Unable to read source file list from python lookup path '%s', skipping\n", path)
			continue
		}

		for _, sourceFile := range sourceFiles {

			fullPath, err = filepath.Abs(sourceFile)
			if(err != nil) {
				continue
			}

			module = filepath.Base(sourceFile)
			module = strings.TrimSuffix(module, filepath.Ext(module))
			ret[module] = fullPath
		}
	}

	return ret
}
