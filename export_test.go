package testinfo

func ExportSetFuncNameFlag(s string) func() {
	old := funcName
	funcName = s
	return func(){
		funcName = old
	}
}