package testinfo

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"path/filepath"
	"strings"
)

const doc = "testinfo is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "testinfo",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

var pref = [...]string {"Test", "Benchmark", "Example"}
type FuncData struct {
	testType    int
	FuncObj     types.Object
	TestDecl    *ast.FuncDecl
	CallPos     []token.Pos
}

func (x FuncData) Type () string {
	return pref[x.testType]
}

type TestInfo struct {
	Pass *analysis.Pass
	FuncData []FuncData
}

func New(pass *analysis.Pass) (TestInfo, error) {
	ti := TestInfo{Pass: pass}
	err := ti.getFuncData(pass)
	if err != nil {
		return ti, err
	}
	return ti, nil
}

func (t *TestInfo) FormatObj(x types.Object) string {
	if x == nil {
		return "unknown"
	}
	fs := t.Pass.Fset
	p := fs.Position(x.Pos())
	s := fmt.Sprintf("\"%s:%s:%d:%d %s\"", x.Pkg().Name(), filepath.Base(p.Filename), p.Line, p.Column, x.Name())
	return s
}

func (t *TestInfo) Format(x FuncData) string {
	testObj := t.Pass.TypesInfo.ObjectOf(x.TestDecl.Name)

	s := fmt.Sprintf(
		"{" +
		strings.Join(
			[]string{
				"type:%s",
				"testFunc:%s",
				"targetFunc:%s",
				"CallPos:["},
				", "),
		x.Type(),
		t.FormatObj(testObj),
		t.FormatObj(x.FuncObj))

	for j, cp := range x.CallPos {
		p := t.Pass.Fset.Position(cp)
		s += fmt.Sprintf("\"%d:%d\"", p.Line, p.Column)
		if j != len(x.CallPos) - 1 {
			s += ", "
		}
	}
	return s + "]}"
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

func getFuncObj(pkg *types.Package, name string) types.Object {
	pName := strings.TrimSuffix(pkg.Name(), "_test")
	if obj := pkg.Scope().Lookup(name); obj != nil {
		return obj
	}
	lower := strings.ToLower(string(name[0]))+name[1:]
	if obj := pkg.Scope().Lookup(lower); obj != nil {
		return obj
	}
	for _, imp := range pkg.Imports() {
		if imp.Name() == pName {
			if obj := imp.Scope().Lookup(name); obj != nil {
				return obj
			}
		}
	}
	return nil
}

func (t *TestInfo) getFuncData(pass *analysis.Pass) error {
	fs := pass.Fset
	files := pass.Files
	info := pass.TypesInfo

	var mainFiles, testFiles []*ast.File
	for _, f := range files {
		path := fs.Position(f.Package).Filename
		if strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, f)
		} else {
			mainFiles = append(mainFiles, f)
		}
	}

	var funcData []FuncData

	for _, f := range testFiles {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			name := fd.Name.String()
			for i, ps := range pref {
				if strings.HasPrefix(name, ps) {
					nameBody := strings.TrimPrefix(name, ps)
					obj := getFuncObj(pass.Pkg, nameBody)
					funcData = append(funcData, FuncData{
						testType:    i,
						FuncObj: obj,
						TestDecl: fd,
						CallPos:     callPosList(fd, obj, info),
					})
				}
			}
		}
	}
	t.FuncData = funcData
	return nil
}

// Posが渡された時に対応する物を返す
func (t *TestInfo) GetFuncDataFromCursor(pos token.Pos) *FuncData {
	for _, x := range t.FuncData {
		if scope, ok := t.Pass.TypesInfo.Scopes[x.TestDecl]; ok && scope.Contains(pos) {
			return &x
		}
	}
	return nil
}

// Posが渡された時に対応する物を返す
func (t *TestInfo) GetFuncDataFromName(funcName string) *FuncData {
	for _, x := range t.FuncData {
		if x.TestDecl.Name.Name == funcName {
			return &x
		}
	}
	return nil
}

var funcName string
func init() {
	Analyzer.Flags.StringVar(&funcName, "testfunc", funcName, "test function name")
}

func run(pass *analysis.Pass) (interface{}, error) {

	testInfo, err := New(pass)
	if err != nil {
		return nil, err
	}

	if funcName == "" {
		for _, x := range testInfo.FuncData {
			pass.Reportf(x.TestDecl.Pos(), testInfo.Format(x))
		}
	} else {
		x := testInfo.GetFuncDataFromName(funcName)
		if x != nil {
			pass.Reportf(x.TestDecl.Pos(), testInfo.Format(*x))
		}
	}

	return nil, nil
}

