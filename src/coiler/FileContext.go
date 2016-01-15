package coiler

import (
	"path/filepath"
)

type FileContext struct {

	context *BuildContext

	// all imported modules, in order.
	dependencies []string

	// The file-local representation of imports, mapped back to fully-qualified function/variable names.
	localSymbols map[string]string

	// the absolute path to the file being combined
	fullPath string

	// the namespace of the file (usually just the name).
	namespace string
}

func NewFileContext(path string, context *BuildContext) (*FileContext, error) {

	var ret *FileContext
	var err error

	ret = new(FileContext)

	ret.fullPath, err = filepath.Abs(path)
	if(err != nil) {
		return nil, err
	}

	ret.context = context
	ret.namespace = filepath.Base(ret.fullPath)
	ret.namespace = ret.namespace[0:len(ret.namespace)-3] // trim *.py extension

	return ret, nil
}
