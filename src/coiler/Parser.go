package coiler

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// import regexes
var standardImportRegex *regexp.Regexp
var aliasedImportRegex *regexp.Regexp
var singleImportRegex *regexp.Regexp
var singleAliasedImportRegex *regexp.Regexp
var wildImportRegex *regexp.Regexp

var classRegex *regexp.Regexp
var functionRegex *regexp.Regexp
var assignmentRegex *regexp.Regexp

func init() {

	standardImportRegex = regexp.MustCompile("import ([a-zA-Z0-9_]+)\\s*$")
	aliasedImportRegex = regexp.MustCompile("import ([a-zA-Z0-9_]+) as ([a-zA-Z0-9_]+)\\s*$")
	singleImportRegex = regexp.MustCompile("from ([a-zA-Z0-9_]+) import ([a-zA-Z0-9_]+)\\s*$")
	singleAliasedImportRegex = regexp.MustCompile("from ([a-zA-Z0-9_]+) import ([a-zA-Z0-9_]+) as ([a-zA-Z0-9_]+)\\s*$")
	wildImportRegex = regexp.MustCompile("from ([a-zA-Z0-9_]+) import \\*\\s*$")

	classRegex = regexp.MustCompile("class ([a-zA-Z0-9_]+).*:")
	functionRegex = regexp.MustCompile("def ([a-zA-Z0-9_]+)\\(.*\\):")
	assignmentRegex = regexp.MustCompile("([a-zA-Z0-9_]+)\\s+=")
}

/*
	Parses the given [inputPath], traverses and processes all dependent imports (combining as required),
	and uses python to compile a real executable to the given [outputPath].
*/
func Parse(inputPath string, useSystemPaths bool) (*BuildContext, error) {

	var context *BuildContext
	var err error

	context = NewBuildContext(useSystemPaths)

	// combine in context
	_, err = parse(inputPath, context)
	if err != nil {
		return nil, err
	}

	return context, nil
}

func parse(path string, context *BuildContext) (*FileContext, error) {

	var fileContext *FileContext
	var contents []byte
	var sourceChannel chan string
	var err error

	sourceChannel = make(chan string)

	contents, err = ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fileContext, err = NewFileContext(path, context)
	if err != nil {
		return nil, err
	}

	context.AddDependency(fileContext)
	go readLines(string(contents), sourceChannel)

	for line := range sourceChannel {

		err = parseLine(line, fileContext, context)
		if err != nil {
			return nil, err
		}
	}

	return fileContext, nil
}

func readLines(source string, output chan string) {

	var newlineIndex int

	defer close(output)

	newlineIndex = -1

	for newlineIndex != len(source) {

		source = source[newlineIndex+1:]
		newlineIndex = strings.Index(source, "\n")

		if newlineIndex < 0 {
			newlineIndex = len(source)
		}

		output <- string(source[0:newlineIndex])
	}
}

func parseLine(line string, fileContext *FileContext, buildContext *BuildContext) error {

	var symbol []string
	var trimmedLine string

	trimmedLine = strings.Trim(line, " \t\r\n")

	// any import
	if strings.Contains(line, "import") {
		parseImport(trimmedLine, fileContext, buildContext)
	}

	// if this line is indented, ignore it. We only need to translate top-level stuff.
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return nil
	}

	// classes
	if strings.Contains(line, "class") {

		symbol = classRegex.FindStringSubmatch(line)
		if len(symbol) > 0 {
			addSymbolToContexts(symbol[1], fileContext, buildContext)
		}
	}

	// function
	if strings.Contains(line, "def") {

		symbol = functionRegex.FindStringSubmatch(line)
		if len(symbol) > 0 {
			addSymbolToContexts(symbol[1], fileContext, buildContext)
		}
	}

	// namespace-level variable
	if strings.Contains(line, "=") {

		symbol = assignmentRegex.FindStringSubmatch(line)
		if len(symbol) > 0 {
			addSymbolToContexts(symbol[1], fileContext, buildContext)
		}
	}

	return nil
}

/*
	Parses a single 'import' statement as it occurs in a source file.
	Modifies the file and build contexts as appropriate.
*/
func parseImport(line string, fileContext *FileContext, buildContext *BuildContext) {

	var dependentContext *FileContext
	var matches []string

	// imports can happen in any number of wacky forms
	// go through from most-to-least specific and try to determine which form is being used,
	// and how to modify the contexts
	matches = wildImportRegex.FindStringSubmatch(line)
	if len(matches) > 0 {
		fmt.Println("Wild import statement detected. Ignoring.")
		return
	}

	matches = singleAliasedImportRegex.FindStringSubmatch(line)
	if len(matches) > 0 {

		dependentContext = parseAndImport(matches[1], fileContext, buildContext)
		if dependentContext == nil {
			return
		}

		fileContext.AliasCall(dependentContext, matches[2], matches[3])
		return
	}

	matches = singleImportRegex.FindStringSubmatch(line)
	if len(matches) > 0 {

		dependentContext = parseAndImport(matches[1], fileContext, buildContext)
		if dependentContext == nil {
			return
		}

		fileContext.UnaliasedCall(dependentContext, matches[2])
		return
	}

	matches = aliasedImportRegex.FindStringSubmatch(line)
	if len(matches) > 0 {

		dependentContext = parseAndImport(matches[1], fileContext, buildContext)
		if dependentContext == nil {
			return
		}

		fileContext.AliasContext(dependentContext, matches[2])
		return
	}

	matches = standardImportRegex.FindStringSubmatch(line)
	if len(matches) > 0 {

		parseAndImport(matches[1], fileContext, buildContext)
		return
	}
}

func parseAndImport(module string, fileContext *FileContext, buildContext *BuildContext) *FileContext {

	var dependentContext *FileContext
	var fullPath string

	if !buildContext.IsFileImported(module) {

		fullPath = buildContext.FindSourcePath(module)
		if fullPath != "" {

			buildContext.AddImportedFile(module)
			dependentContext, _ = parse(fullPath, buildContext)
			fileContext.AddDependency(module)
		} else {
			buildContext.AddExternalDependency(module)
		}
	} else {
		dependentContext = buildContext.GetFileContext(module)
	}

	return dependentContext
}

/*
	Properly adds the given [symbol] to the given file and build contexts.
*/
func addSymbolToContexts(symbol string, fileContext *FileContext, buildContext *BuildContext) {

	var qualifiedName string

	qualifiedName = fileContext.AddLocalSymbol(symbol)
	buildContext.AddSymbol(qualifiedName)
}
