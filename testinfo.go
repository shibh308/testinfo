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
	PackageName string
	FuncName    string
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
	DeclList map[string]map[string]*ast.FuncDecl
	TestList map[string]map[string]*ast.FuncDecl
}

func New(pass *analysis.Pass) (TestInfo, error) {
	ti := TestInfo{Pass: pass}
	err := ti.getFuncData(pass)
	if err != nil {
		return ti, err
	}
	return ti, nil
}

func Format(x FuncData, fs *token.FileSet) string {

	var fStr string
	if x.FuncObj == nil {
		fStr = "FuncDeclPos:unknown"
	} else {
		fStr = fmt.Sprintf("FuncObjPos:\"%s,%d,%d\"", filepath.Base(fs.Position(x.FuncObj.Pos()).Filename), fs.Position(x.FuncObj.Pos()).Line, fs.Position(x.FuncObj.Pos()).Column)
	}
	s := fmt.Sprintf(
		"{" +
		strings.Join(
			[]string{
				"PackageName:%s",
				"type:%s",
				"FuncName:%s",
				fStr,
				"TestDeclPos:\"%s,%d,%d\"",
				"CallPos:["}, ", "),
		x.PackageName, x.Type(), x.FuncName,
		filepath.Base(fs.Position(x.TestDecl.Pos()).Filename), fs.Position(x.TestDecl.Pos()).Line, fs.Position(x.TestDecl.Pos()).Column,
	)
	for j, cp := range x.CallPos {
		s += fmt.Sprintf("\"%s,%d,%d\"", filepath.Base(fs.Position(cp).Filename), fs.Position(cp).Line, fs.Position(cp).Column)
		if j != len(x.CallPos) - 1 {
			s += ", "
		}
	}
	return s + "]}"
}

func callPosList(n *ast.FuncDecl, target types.Object, info *types.Info) []token.Pos {
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

func (testInfo *TestInfo) getFuncData(pass *analysis.Pass) error {
	fs := pass.Fset
	files := pass.Files
	info := pass.TypesInfo
	pName := strings.TrimSuffix(pass.Pkg.Name(), "_test")

	var mainFiles, testFiles []*ast.File
	for _, f := range files {
		path := fs.Position(f.Package).Filename
		if strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, f)
		} else {
			mainFiles = append(mainFiles, f)
		}
	}

	testMap := make(map[string]map[string]*ast.FuncDecl)

	/*
	declMap := make(map[string]map[string]*ast.FuncDecl)
	for _, f := range mainFiles {
		pName := f.Name.String()
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			fName := strings.ToLower(fd.Name.String())
			if _, exists := declMap[pName]; !exists {
				declMap[pName] = make(map[string]*ast.FuncDecl)
			}
			declMap[pName][fName] = fd
		}
	}
	 */

	var funcData []FuncData

	for _, f := range testFiles {
		/*
		if _, ok := declMap[pName]; !ok {
			continue
		}
		 */
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			name := fd.Name.String()
			for i, ps := range pref {
				if strings.HasPrefix(name, ps) {
					nameBody := strings.TrimPrefix(name, ps)

					obj := pass.Pkg.Scope().Lookup(nameBody)
					if obj == nil {
						lower := strings.ToLower(string(nameBody[0]))+nameBody[1:]
						obj = pass.Pkg.Scope().Lookup(lower)
					}
					for _, pkg := range pass.Pkg.Imports() {
						if pkg.Name() == pName {
							obj = pkg.Scope().Lookup(nameBody)
						}
					}

					_, _ = info, i
					if obj != nil {
						if _, ok := testMap[pName]; !ok {
							testMap[pName] = make(map[string]*ast.FuncDecl)
						}
						testMap[pName][strings.ToLower(nameBody)] = fd
						funcData = append(funcData, FuncData{
							PackageName: pName,
							FuncName:    nameBody,
							testType:    i,
							FuncObj: obj,
							TestDecl: fd,
							CallPos:     callPosList(fd, obj, info),
							// CallPos:     callPosList(fd, info.Defs[decl.Name], info),
						})
					} else {
						funcData = append(funcData, FuncData{
							PackageName: pName,
							FuncName:    nameBody,
							testType:    i,
							FuncObj: nil,
							TestDecl: fd,
							CallPos:  nil,
						})
					}
				}
			}
		}
	}
	testInfo.FuncData = funcData
	// testInfo.DeclList = declMap
	testInfo.TestList = testMap
	return nil
}

// Posが渡された時に対応する物を返す
func (testInfo *TestInfo) GetCursorFuncData(pos token.Pos) *FuncData {
	for _, x := range testInfo.FuncData {
		/*
		if scope, ok := testInfo.Pass.TypesInfo.Scopes[x.FuncObj.]; ok && scope.Contains(pos) {
			return &x
		}
		 */
		if scope, ok := testInfo.Pass.TypesInfo.Scopes[x.TestDecl]; ok && scope.Contains(pos) {
			return &x
		}
	}
	return nil
}

func run(pass *analysis.Pass) (interface{}, error) {

	testInfo, err := New(pass)
	if err != nil {
		return nil, err
	}

	for _, x := range testInfo.FuncData {
		pass.Reportf(x.TestDecl.Pos(), Format(x, pass.Fset))
	}

	return nil, nil
}

