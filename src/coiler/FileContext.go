package coiler

import (
	"path/filepath"
	"strings"
)

type FileContext struct {

	context *BuildContext

	// all imported modules, in order.
	dependencies []string

	// The file-local representation of imports, mapped back to fully-qualified function/variable names.
	localSymbols map[string]string

	// the imported symbols (that may or may not be aliased), key is the as-found-in-source symbol name,
	// value is the fully-qualified function/variable name
	dependentSymbols map[string]string

	// the absolute path to the file being combined
	fullPath string

	// the namespace of the file (usually just the name).
	namespace string
}

func NewFileContext(path string, context *BuildContext) (*FileContext, error) {

	var ret *FileContext
	var err error

	ret = new(FileContext)
	ret.localSymbols = make(map[string]string)
	ret.dependentSymbols = make(map[string]string)

	ret.fullPath, err = filepath.Abs(path)
	if(err != nil) {
		return nil, err
	}

	ret.context = context
	ret.namespace = filepath.Base(ret.fullPath)
	ret.namespace = ret.namespace[0:len(ret.namespace)-3] // trim *.py extension

	return ret, nil
}

/*
	Translates a single line of source using the local symbol map.
*/
func (this *FileContext) TranslateLine(line string) string {

	for key, value := range this.localSymbols {
		line = strings.Replace(line, key, this.context.TranslateSymbol(value), -1)
	}

	return line
}

/*
	Adds all symbols of the given [dependentContext] to this local symbol table,
	prefixing them with the given [alias]
*/
func (this *FileContext) AliasContext(dependentContext *FileContext, alias string) {

	var aliasedSymbol string
	var bareSymbol string

	for _, fullSymbol := range dependentContext.localSymbols {

		bareSymbol = strings.Replace(fullSymbol, dependentContext.namespace + ".", "", -1)
		aliasedSymbol = alias + "." + bareSymbol

		this.dependentSymbols[aliasedSymbol] = fullSymbol
	}
}

func (this *FileContext) AddDependency(module string) {
	this.dependencies = append(this.dependencies, module)
}

func (this *FileContext) AddLocalSymbol(localName string) string {

	var qualifiedName string

	qualifiedName = this.namespace + "." + localName
	this.localSymbols[localName] = qualifiedName

	return qualifiedName
}
