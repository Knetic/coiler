package coiler

/*
	Handles the compilation to a binary embedded executable.
	This is accomplished by writing a bootstrap program which runs a python interpreter,
	taking the source for that interpreter from the same file as the executable (the pyc code is appendedto the end of the executable)
*/
import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"fmt"
	"os"
	"io"
	"errors"
	"os/exec"
)

const (
	EMBEDDED_SOURCE = `
#include <stdio.h>

void main(const int arc, const char** arv)
{
		printf("Hello, embedded world!\n");
}
`

	MAGIC_STRING = "COILER_SRC:"
)

var NUL_ENTRIES = []byte {
	0, 0, 0, 0, 0,
	0, 0, 0, 0, 0,
	0, 0, 0, 0, 0,
	0, 0, 0, 0, 0,
	0, 0, 0,
}

func CreateBinary(sourcePath string) error {

	var compiledPath string
	var precompiledPath string
	var baseName string
	var err error

	precompiledPath, err = ioutil.TempDir("", "coilerEmbedded")
	if err != nil {
		return err
	}

	baseName = filepath.Base(sourcePath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	precompiledPath = filepath.Join(precompiledPath, (baseName + ".c"))
	compiledPath, err = filepath.Abs(sourcePath)
	if(err != nil) {
		return err
	}

	compiledPath = filepath.Dir(compiledPath)
	compiledPath = filepath.Join(compiledPath, baseName)

	err = writeEmbeddedSource(precompiledPath)
	if(err != nil) {
		return err
	}

 	err = compileEmbedded(precompiledPath, compiledPath)
	if(err != nil) {
		return err
	}

	err = appendApplication(sourcePath, compiledPath)
	return err
}

func writeEmbeddedSource(target string) error {

	err := ioutil.WriteFile(target, []byte(EMBEDDED_SOURCE), 0644)
	return err
}

func compileEmbedded(sourcePath string, targetPath string) error {

	var compiler *exec.Cmd
	var arguments []string
	var rawOutput []byte
	var err error

	arguments = []string {
		"-o",
		targetPath,
		sourcePath,
	}

	compiler = exec.Command("gcc", arguments...)
	rawOutput, err = compiler.CombinedOutput()

	if(err != nil) {
		errorMsg := fmt.Sprintf("%v\n%v\n", err.Error(), string(rawOutput))
		return errors.New(errorMsg)
	}

	return nil
}

/*
	Appends the *.pyc code at the given [source] to the end of the given [target].
	But first, appends 23 NUL characters followed by a magic string (as identifiers).
*/
func appendApplication(source string, target string) error {

	var sourceFile, targetFile *os.File
	var err error

	targetFile, err = os.OpenFile(target, os.O_APPEND | os.O_WRONLY, 0755)
	if(err != nil) {
		return err
	}
	defer targetFile.Close()

	sourceFile, err = os.OpenFile(source, os.O_RDONLY, 0644)
	if(err != nil) {
		return err
	}
	defer sourceFile.Close()

	targetFile.Write(NUL_ENTRIES)
	targetFile.WriteString(MAGIC_STRING)

	_, err = io.Copy(targetFile, sourceFile)
	return err
}
