package coiler

import (
	"path/filepath"
	"strings"
	"regexp"
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

var invalidPythonCharacters *regexp.Regexp
var irreplaceableCharacters *regexp.Regexp

func init() {

	invalidPythonCharacters = regexp.MustCompile("[^a-zA-Z0-9_]")
	irreplaceableCharacters = regexp.MustCompile("[a-zA-Z0-9_]")
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
	ret.namespace = invalidPythonCharacters.ReplaceAllString(ret.namespace, "")

	return ret, nil
}

/*
	Translates a single line of source using the local symbol map.
*/
func (this *FileContext) TranslateLine(line string) string {

	var keys []string

	// ignore any top-level imports (but leave imports that are mid-line, since they're probably conditional)
	if(strings.HasPrefix(line, "import") || (strings.HasPrefix(line, "from") && strings.Contains(line, "import"))) {
		return ""
	}

	// TODO: need to make sure none of these replacements happen between valid python alphanumeric characters
	keys = orderMapKeysByLength(this.localSymbols)
	for _, key := range keys {
		replaceSymbol(line, key, this.context.TranslateSymbol(this.localSymbols[key]))
	}

	keys = orderMapKeysByLength(this.dependentSymbols)
	for _, key := range keys {
		replaceSymbol(line, key, this.context.TranslateSymbol(this.dependentSymbols[key]))
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

func (this *FileContext) AliasCall(dependentContext *FileContext, remoteName string, aliasedName string) {

	var fullSymbol string

	fullSymbol = dependentContext.namespace + "." + remoteName
	this.dependentSymbols[aliasedName] = fullSymbol
}

func (this *FileContext) UnaliasedCall(dependentContext *FileContext, remoteName string) {

	var fullSymbol string

	fullSymbol = dependentContext.namespace + "." + remoteName
	this.dependentSymbols[remoteName] = fullSymbol
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

/*
	Returns a list of keys where the longest keys are given first, shortest last.
*/
func orderMapKeysByLength(source map[string]string) []string {

	var ret []string
	var swap string
	var length int
	var index int

	length = len(source)
	ret = make([]string, length)
	index = 0

	// first just make the array
	for key, _ := range source {
		ret[index] = key
		index++
	}

	// BUBBLE SORT!
	for i := 0; i < length; i++ {
		for z := 0; z < length-1; z++ {

			if(len(ret[z]) < len(ret[z+1])) {
				swap = ret[z]
				ret[z] = ret[z+1]
				ret[z+1] = swap
			}
		}
	}

	return ret
}

/*
	Replaces all occurrences of [symbol] in the given [line], as long as
	it is not surrounded by valid python alphanumeric identifiers.
*/
func replaceSymbol(line string, symbol string, replacement string) string {

	var prefix, postfix []byte
	var startIndex, endIndex int

	endIndex = -1

	for {

		endIndex++
		if(endIndex > len(line)) {
			return line
		}

		startIndex = strings.Index(line[endIndex:], symbol)
		if(startIndex < 0) {
			return line
		}

		// TODO: make sure string literals (and comments?) don't get translated

		startIndex += endIndex
		endIndex = startIndex + len(symbol)

		// if prefix is within range
		if(startIndex-1 > 0) {

			// if prefixed by an alphanumeric character
			prefix = []byte(line[startIndex-1:startIndex])
			if(irreplaceableCharacters.Match(prefix)) {

				//fmt.Printf("Found a prefix replacement (%s) that shouldn't happen: line '%s', symbol: '%s'\n%d:%d\n\n", prefix, line, symbol, startIndex, endIndex)
				continue
			}
		}

		// if postfix is within range
		if(endIndex < len(line)) {

			// if postfixed by an alphanumeric character
			postfix = []byte(line[endIndex:endIndex+1])
			if(irreplaceableCharacters.Match(postfix)) {

				//fmt.Printf("Found a postfix (%s) replacement that shouldn't happen: line '%s', symbol: '%s'\n%d:%d\n\n", postfix, line, symbol, startIndex, endIndex)
				continue
			}
		}

		// replace
		line = line[0:startIndex] + replacement + line[endIndex:]
		endIndex -= (len(symbol) - len(replacement))
	}
	return line
}
