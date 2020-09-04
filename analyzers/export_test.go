package analyzers

func ExportSetFuncNameFlag(s string) func() {
	old := flags.funcName
	flags.funcName = s
	return func(){
		flags.funcName = old
	}
}

func ExportSetFileNameFlag(s string) func() {
	old := flags.fileName
	flags.fileName = s
	return func(){
		flags.fileName = old
	}
}
