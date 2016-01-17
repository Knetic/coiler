package coiler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CompileCombinedFile(outputPath string, context *BuildContext) error {

	var precompiledOutputPath string
	var compiledName, precompiledName, baseName string
	var err error

	precompiledOutputPath, err = ioutil.TempDir("", "coiler")
	if err != nil {
		return err
	}

	baseName = filepath.Base(outputPath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	precompiledName = fmt.Sprintf("%s/%s.py", precompiledOutputPath, baseName)

	if strings.HasSuffix(baseName, ".pyc") {
		compiledName = fmt.Sprintf("%s/%s", precompiledOutputPath, baseName)
	} else {
		compiledName = fmt.Sprintf("%s/%s.pyc", precompiledOutputPath, baseName)
	}

	err = writeCombinedOutput(precompiledName, context)
	if err != nil {
		return err
	}

	err = callPythonCompiler(precompiledOutputPath)
	if err != nil {
		return err
	}

	err = copyFile(compiledName, outputPath)
	return err
}

func callPythonCompiler(targetPath string) error {

	var compiler *exec.Cmd
	var arguments []string
	var output []byte
	var err error

	arguments = []string{"-m", "compileall", targetPath}

	compiler = exec.Command("python", arguments...)

	fmt.Println("Calling python compiler")
	output, err = compiler.CombinedOutput()

	if err != nil {
		errorMsg := fmt.Sprintf("Compile failed:\n%s\n%v\n", string(output), err)
		return errors.New(errorMsg)
	}
	return nil
}

/*
	Brute copy from [source] to [target].
*/
func copyFile(source, target string) error {

	var sourceFile, targetFile *os.File
	var err error

	sourceFile, err = os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err = os.Create(target)
	if err != nil {
		return err
	}
	targetFile.Chmod(0755)
	defer targetFile.Close()

	_, err = io.Copy(sourceFile, targetFile)
	return err
}

/*
	Takes the current build context and writes a single combined source file to the given [targetPath]
*/
func writeCombinedOutput(targetPath string, buildContext *BuildContext) error {

	var fileContexts []*FileContext
	var outFile *os.File
	var line string
	var err error

	outFile, err = os.Create(targetPath)
	if err != nil {

		outFile, err = os.Open(targetPath)
		if err != nil {
			return err
		}
	}
	defer outFile.Close()

	// write external dependencies first
	for _, dependency := range buildContext.externalDependencies {

		line = fmt.Sprintf("import %v\n", dependency)
		outFile.Write([]byte(line))
	}

	buildContext.dependencies.DiscoverNeighbors()
	fileContexts = buildContext.dependencies.GetOrderedNodes()

	for _, context := range fileContexts {

		err = writeTranslatedFile(context, outFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeTranslatedFile(context *FileContext, outFile *os.File) error {

	var sourceFile *os.File
	var sourceReader *bufio.Reader
	var line string
	var rawLine []byte
	var err error

	sourceFile, err = os.Open(context.fullPath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	sourceReader = bufio.NewReader(sourceFile)

	for {
		rawLine, err = sourceReader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = context.TranslateLine(string(rawLine))
		outFile.Write([]byte(line))
	}
	return nil
}
