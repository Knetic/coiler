package coiler

type FileCombiner struct {

	context *BuildContext

	// The file-local representation of imports, mapped back to fully-qualified function/variable names.
	localSymbols map[string]string

	// the absolute path to the file being combined
	fullPath string

	// the namespace of the file (usually just the name).
	namespace string
}
