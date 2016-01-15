package coiler

import (
	"io"
	"io/ioutil"
	"path"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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

	fmt.Printf("Line: %s\n", line)
	return nil
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
