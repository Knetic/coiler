package coiler

import (
	"io"
	"io/ioutil"
	"path"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"regexp"
)

var standardImportRegex *regexp.Regexp
var aliasedImportRegex *regexp.Regexp
var singleImportRegex *regexp.Regexp
var singleAliasedImportRegex *regexp.Regexp
var wildImportRegex *regexp.Regexp

func init() {

	standardImportRegex = regexp.MustCompile("import ([a-zA-Z0-9_]+)")
	aliasedImportRegex = regexp.MustCompile("import ([a-zA-Z0-9_]+) as ([a-zA-Z0-9_]+)")
	singleImportRegex = regexp.MustCompile("from ([a-zA-Z0-9_]+) import ([a-zA-Z0-9_]+)")
	singleAliasedImportRegex = regexp.MustCompile("from ([a-zA-Z0-9_]+) import ([a-zA-Z0-9_]+) as ([a-zA-Z0-9_]+)")
	wildImportRegex = regexp.MustCompile("from ([a-zA-Z0-9_]+) import \\*")
}

/*
	Parses the given [inputPath], traverses and processes all dependent imports (combining as required),
	and uses python to compile a real executable to the given [outputPath].
*/
func Parse(inputPath string, outputPath string) error {

	var context *BuildContext
	var precompiledOutputPath string
	var compiledName, precompiledName, baseName string
	var err error

	context = NewBuildContext()

	// combine in context
	err = parse(inputPath, context)
	if(err != nil) {
		return err
	}

	precompiledOutputPath, err = ioutil.TempDir("", "coiler")
	if(err != nil) {
		return err
	}

	baseName = path.Base(outputPath)
	precompiledName = fmt.Sprintf("%s/%s.py", precompiledOutputPath, baseName)

	if(strings.HasSuffix(baseName, ".pyc")) {
		compiledName = fmt.Sprintf("%s/%s", precompiledOutputPath, baseName)
	} else {
		compiledName = fmt.Sprintf("%s/%s.pyc", precompiledOutputPath, baseName)
	}

	err = context.WriteCombinedOutput(precompiledName)
	if(err != nil) {
		return err
	}

	err = callPythonCompiler(precompiledOutputPath)
	if(err != nil) {
		return err
	}

	err = copyFile(compiledName, outputPath)
	return err
}

func parse(path string, context *BuildContext) error {

	var fileContext *FileContext
	var contents []byte
	var sourceChannel chan string
	var err error

	sourceChannel = make(chan string)

	contents, err = ioutil.ReadFile(path)
	if(err != nil) {
		return err
	}

	fileContext, err = NewFileContext(path, context)
	if(err != nil) {
		return err
	}

	go readLines(string(contents), sourceChannel)

	for line := range sourceChannel {
		err = parseLine(line, fileContext, context)
		if(err != nil) {
			return err
		}
	}

	return nil
}

func readLines(source string, output chan string) {

	var newlineIndex int

	defer close(output)

	newlineIndex = -1

	for newlineIndex != len(source) {

		source = source[newlineIndex+1:]
		newlineIndex = strings.Index(source, "\n")

		if(newlineIndex < 0) {
			newlineIndex = len(source)
		}

		output <- string(source[0:newlineIndex])
	}
}

func parseLine(line string, fileContext *FileContext, buildContext *BuildContext) error {

	line = strings.Trim(line, " \t\n\r")

	if(strings.Contains(line, "import")) {
		parseImport(line, fileContext, buildContext)
	}
	return nil
}

/*
	Parses a single 'import' statement as it occurs in a source file.
	Modifies the file and build contexts as appropriate.
*/
func parseImport(line string, fileContext *FileContext, buildContext *BuildContext) {

	//var module, alias, individual string
	var matches []string

	// imports can happen in any number of wacky forms
	// go through from most-to-least specific and try to determine which form is being used,
	// and how to modify the contexts
	matches = wildImportRegex.FindStringSubmatch(line)
	if(len(matches) > 0) {
		fmt.Println("Wild import statement detected. Ignoring.")
		return
	}

	matches = singleAliasedImportRegex.FindStringSubmatch(line)
	if(len(matches) > 0) {
		fmt.Printf("Found module import '%s', function '%s' aliased to '%s'\n", matches[1], matches[2], matches[3])
		return
	}

	matches = singleImportRegex.FindStringSubmatch(line)
	if(len(matches) > 0) {
		fmt.Printf("Found module import '%s', function '%s'\n", matches[1], matches[2])
		return
	}

	matches = aliasedImportRegex.FindStringSubmatch(line)
	if(len(matches) > 0) {
		fmt.Printf("Found aliased module import '%s', aliased to '%s'\n", matches[1], matches[2])
		return
	}

	matches = standardImportRegex.FindStringSubmatch(line)
	if(len(matches) > 0) {
		fmt.Printf("Found normal import '%s'\n", matches[1])
		return
	}
}

func callPythonCompiler(targetPath string) error {

	var compiler *exec.Cmd
	var arguments []string

	arguments = []string {"-m", "compileall", targetPath}

	compiler = exec.Command("python", arguments...)
	return compiler.Run()
}

/*
	Brute copy from [source] to [target].
*/
func copyFile(source, target string) error {

	var sourceFile, targetFile *os.File
	var err error

	sourceFile, err = os.Open(source)
	if(err != nil) {
		return err
	}
	defer sourceFile.Close()

	targetFile, err = os.Create(target)
	if(err != nil) {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(sourceFile, targetFile)
	return err
}
