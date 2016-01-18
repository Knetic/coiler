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
#include <stdlib.h>
#include <string.h>
#include <Python.h>

int memsearch(const char *hay, int haysize, const char *needle, int needlesize) {
    int haypos, needlepos;
    haysize -= needlesize;
    for (haypos = 0; haypos <= haysize; haypos++) {
        for (needlepos = 0; needlepos < needlesize; needlepos++) {
            if (hay[haypos + needlepos] != needle[needlepos]) {
                // Next character in haystack.
                break;
            }
        }
        if (needlepos == needlesize) {
            return haypos;
        }
    }
    return -1;
}

/*
	Searches for the magic string, returning the byte offset at which data can be read after the magic string.
	-1 indicates no magic string found.
*/
long extractPYC(FILE* sourceFile)
{
	char buffer[256];
	char searchString[30];
	size_t subIndex;
	size_t frontBuffer;
	long offset;
	int bytesRead;

	frontBuffer = buffer + 128;
	subIndex = -1;
	offset = 0;

	// form the magic string to search for.
	for(int i = 0; i < 23; i++)
		searchString[i] = '\0';
	memcpy(searchString + 23, "COILER:", 7);

	while(1)
	{
		// move front half to the back
		memcpy(buffer, frontBuffer, 128);

		// read data to front
		bytesRead = fread(frontBuffer, 1, 128, sourceFile);

		// eof?
		if(bytesRead <= 0)
			break;

		subIndex = memsearch(buffer, 256, searchString, 30);
		if(subIndex != -1)
		{
			offset += subIndex - 128;
			return offset;
		}

		offset += bytesRead;
	}

	return -1;
}

char *gnu_basename(char *path)
{
    char *base = strrchr(path, '/');
    return base ? base+1 : path;
}

int main(const int arc, const char** argv)
{
	FILE* executable;
	long applicationOffset;
	int status;

	executable = fopen(argv[0], "r");
	if(executable == 0)
	{
		printf("Unable to read own executable\n");
		return 1;
	}

	applicationOffset = extractPYC(executable);
	fclose(executable);

	if(applicationOffset < 0)
	{
		printf("Unable to find PYC application code\n");
		return 1;
	}

	executable = fopen(argv[0], "r");
	status = fseek(executable, applicationOffset, 0);
	if(status != 0)
	{
		printf("Unable to skip to pyc application\n");
		return 1;
	}

	Py_SetProgramName(argv[0]);
	Py_Initialize();
	status = PyRun_SimpleFile(executable, gnu_basename(argv[0]));
	Py_Finalize();
	return status;
}
`

	MAGIC_STRING = "COILER:"
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
	var output string
	var err error

	// first, call pythonX.Y-config to get list of includes needed into order to embed python
	compiler = exec.Command("python2.7-config", "--cflags")
	rawOutput, err = compiler.Output()

	if(err != nil) {
		errorMsg := fmt.Sprintf("%v\n%v\n", err.Error(), string(rawOutput))
		return errors.New(errorMsg)
	}
	output = string(rawOutput)
	output = strings.Replace(output, "\n", "", -1)
	output = strings.Replace(output, "  ", " ", -1)
	arguments = strings.Split(output, " ")

	// set up output
	arguments = append(arguments, "-o")
	arguments = append(arguments, targetPath)
	arguments = append(arguments, sourcePath)

	// linker arguments
	compiler = exec.Command("python2.7-config", "--ldflags")
	rawOutput, err = compiler.Output()

	if(err != nil) {
		errorMsg := fmt.Sprintf("%v\n%v\n", err.Error(), string(rawOutput))
		return errors.New(errorMsg)
	}
	output = string(rawOutput)
	output = strings.Replace(output, "\n", "", -1)
	output = strings.Replace(output, "  ", " ", -1)
	arguments = append(arguments, strings.Split(output, " ")...)

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

	sourceFile, err = os.Open(source)
	if(err != nil) {
		return err
	}
	defer sourceFile.Close()

	targetFile.Write(NUL_ENTRIES)
	targetFile.WriteString(MAGIC_STRING)

	_, err = io.Copy(targetFile, sourceFile)
	return err
}
