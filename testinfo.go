package testinfo

import (
	"fmt"
	"go/ast"
	"go/importer"
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
	FuncDeclPos     token.Pos
	TestDeclPos     token.Pos
	CallPos         []token.Pos
}

func (fd FuncData) Type () string {
	return pref[fd.testType]
}

func callPosList (n *ast.FuncDecl, target types.Object, info *types.Info) []token.Pos {
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

func run(pass *analysis.Pass) (interface{}, error) {
	fs := pass.Fset

	cfg := &types.Config{Importer: importer.Default()}
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}, Defs: map[*ast.Ident]types.Object{}}
	_, err := cfg.Check("main", fs, pass.Files, info)
	if err != nil {
		return nil, err
	}

	var mainFiles, testFiles []*ast.File
	for _, f := range pass.Files {
		path := fs.Position(f.Package).Filename
		if strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, f)
		} else {
			mainFiles = append(mainFiles, f)
		}
	}
	declMap := make(map[string]map[string]*ast.FuncDecl)
	for _, f := range mainFiles {
		pName := f.Name.String()
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			fName := fd.Name.String()
			if _, exists := declMap[pName]; !exists {
				declMap[pName] = make(map[string]*ast.FuncDecl)
			}
			declMap[pName][fName] = fd
		}
	}

	var result []FuncData

	for _, f := range testFiles {
		pName := strings.TrimSuffix(f.Name.String(), "_test")
		if _, ok := declMap[pName]; !ok {
			continue
		}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			name := fd.Name.String()
			for i, ps := range pref {
				if strings.HasPrefix(name, ps) {
					nameBody := strings.TrimPrefix(name, ps)
					if decl, ok := declMap[pName][nameBody]; ok {
						x := FuncData{
							PackageName: pName,
							FuncName:    nameBody,
							testType:    i,
							FuncDeclPos: decl.Pos(),
							TestDeclPos: fd.Pos(),
							CallPos:     callPosList(fd, info.Defs[decl.Name], info),
						}
						result = append(result, x)
						s := fmt.Sprintf(
							"{\n" +
								"PackageName: %s\n" +
								"FuncName: %s\n" +
								"type: %s\n" +
								"FuncDeclPos: %s,%d,%d\n" +
								"TestDeclPos: %s,%d,%d\n" +
								"CallPos: [",
							x.PackageName, x.FuncName, x.Type(),
							filepath.Base(fs.Position(x.FuncDeclPos).Filename), fs.Position(x.FuncDeclPos).Line, fs.Position(x.FuncDeclPos).Column,
							filepath.Base(fs.Position(x.TestDeclPos).Filename), fs.Position(x.TestDeclPos).Line, fs.Position(x.TestDeclPos).Column,
						)
						for j, cp := range x.CallPos {
							s += fmt.Sprintf("%s,%d,%d", filepath.Base(fs.Position(cp).Filename), fs.Position(cp).Line, fs.Position(cp).Column)
							if j != len(x.CallPos) - 1 {
								s += ", "
							}
						}
						pass.Reportf(fd.Pos(), s + "]\n}")
					}
				}
			}
		}
	}

	return nil, nil
}

