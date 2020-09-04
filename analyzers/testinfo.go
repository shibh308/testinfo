// analyzers shows the location of test function, target function data, and additional information.
package analyzers

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"strings"
)

var pref = [...]string{"Test", "Benchmark", "Example"}

type FuncData struct {
	testType int
	FuncObj  types.Object
	TestDecl *ast.FuncDecl
	CallPos  []token.Pos
}

// test function type (one of the {"Test", "Benchmark", "Example"})
func (x FuncData) Type() string {
	return pref[x.testType]
}

type TestInfo struct {
	Pass     *analysis.Pass
	FuncData []*FuncData
}

func New(pass *analysis.Pass, filter func(string) bool) (TestInfo, error) {
	ti := TestInfo{Pass: pass}
	err := ti.getFuncData(pass, filter)
	if err != nil {
		return ti, err
	}
	return ti, nil
}

type JsonPos struct {
	Line int
	Col  int `json:"Column"`
	Ofs  int `json:"OffSet"`
}

type JsonObj struct {
	Pkg  string `json:"Package"`
	File string `json:"FilePath"`
	Name string `json:"FuncName"`
	Pos  JsonPos
}

type JsonFunc struct {
	Ty       string   `json:"Type"`
	TestFn   *JsonObj `json:"testFn,omitempty"`
	TargetFn *JsonObj `json:"targetFn,omitempty"`
	CallPos  []JsonPos
}

func (t *TestInfo) FormatPos(pos token.Pos) JsonPos {
	p := t.Pass.Fset.Position(pos)
	return JsonPos{
		p.Line,
		p.Column,
		p.Offset,
	}
}

// instead of String()
func (t *TestInfo) FormatObj(x types.Object) *JsonObj {
	if x == nil {
		return nil
	}
	return &JsonObj{
		x.Pkg().Name(),
		t.Pass.Fset.File(x.Pos()).Name(),
		x.Name(),
		t.FormatPos(x.Pos()),
	}
}

// instead of String()
func (t *TestInfo) Format(x FuncData) *JsonFunc {
	testObj := t.Pass.TypesInfo.ObjectOf(x.TestDecl.Name)

	var callPos []JsonPos
	for _, cp := range x.CallPos {
		callPos = append(callPos, t.FormatPos(cp))
	}
	return &JsonFunc{
		Ty:       x.Type(),
		TestFn:   t.FormatObj(testObj),
		TargetFn: t.FormatObj(x.FuncObj),
		CallPos:  callPos,
	}
}

func callPosList(n *ast.FuncDecl, target types.Object, info *types.Info) []token.Pos {
	if target == nil {
		return nil
	}
	var result []token.Pos
	ast.Inspect(n.Body, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if info.Uses[id] == target {
			result = append(result, id.NamePos)
		}
		return true
	})
	return result
}

func objFuncCheck(obj types.Object) bool {
	if obj == nil {
		return false
	}
	_, ok := obj.(*types.Func)
	return ok
}

func getFuncObj(pkg *types.Package, name string) types.Object {
	pName := strings.TrimSuffix(pkg.Name(), "_test")
	if obj := pkg.Scope().Lookup(name); objFuncCheck(obj) {
		return obj
	}
	lower := strings.ToLower(string(name[0])) + name[1:]
	if obj := pkg.Scope().Lookup(lower); objFuncCheck(obj) {
		return obj
	}
	for _, imp := range pkg.Imports() {
		if imp.Name() == pName {
			if obj := imp.Scope().Lookup(name); objFuncCheck(obj) {
				return obj
			}
		}
	}
	return nil
}

func (t *TestInfo) getFuncData(pass *analysis.Pass, filter func(string) bool) error {
	fs := pass.Fset
	files := pass.Files
	info := pass.TypesInfo

	var funcData []*FuncData

	for _, f := range files {
		path := fs.Position(f.Name.Pos()).Filename
		if !strings.HasSuffix(path, "_test.go") || !filter(path) {
			continue
		}
		fmt.Println(path)
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			name := fd.Name.String()
			for i, ps := range pref {
				if strings.HasPrefix(name, ps) {
					nameBody := strings.TrimPrefix(name, ps)
					if nameBody[0] == '_' {
						nameBody = nameBody[1:]
					}
					obj := getFuncObj(pass.Pkg, nameBody)
					funcData = append(funcData, &FuncData{
						testType: i,
						FuncObj:  obj,
						TestDecl: fd,
						CallPos:  callPosList(fd, obj, info),
					})
				}
			}
		}
	}
	t.FuncData = funcData
	return nil
}

// returns the function which contains cursor
func (t *TestInfo) GetFuncDataFromCursor(pos token.Pos) *FuncData {
	for _, x := range t.FuncData {
		if scope, ok := t.Pass.TypesInfo.Scopes[x.TestDecl]; ok && scope.Contains(pos) {
			return x
		}
	}
	return nil
}

// returns the function which name is funcName
func (t *TestInfo) GetFuncDataFromName(funcName string) *FuncData {
	for _, x := range t.FuncData {
		if x.TestDecl.Name.Name == funcName {
			return x
		}
	}
	return nil
}
