package analysis

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"path/filepath"
	"strings"
)

type FuncData struct {
	Package  string
	FilePath string
	FuncName string
	Offset   int
}

type TestFuncData struct {
	FuncData
	TestType string
}

// testにいない時に返したい値
type MainFuncJson struct {
	Type       string
	TargetFunc FuncData
	TestFunc   *TestFuncData `json:"TestFunc,omitempty"`
}

// testにいる時に返したい値
type TestFuncJson struct {
	Type        string
	TestFunc    TestFuncData
	TargetFunc  *FuncData `json:"TargetFunc,omitempty"`
	SubTestName []string  `json:"SubTestName,omitempty"`
}

type Analysis struct {
	fs      *token.FileSet
	path    string
	offset  int
	pkgs    []*packages.Package
	pkgIdx  int
	fileIdx int
}

func New(ctx context.Context, path string, offset int) (Analysis, error) {
	fs := token.NewFileSet()
	absPath, err := filepath.Abs(path)
	if err != nil {
		return Analysis{}, fmt.Errorf("failed to get absptah: %w", err)
	}

	pkgMode := packages.NeedName |
		packages.NeedFiles |
		packages.NeedImports |
		packages.NeedTypesInfo |
		packages.NeedTypes |
		packages.NeedSyntax

	cfg := &packages.Config{
		Mode:    pkgMode,
		Context: ctx,
		Dir:     filepath.Dir(path),
		Fset:    fs,
		Tests:   true,
	}

	pkgs, err := packages.Load(cfg)
	if err != nil {
		return Analysis{}, fmt.Errorf("failed to load package: %w", err)
	}

	var pkgIdx, fileIdx int
	var selected bool
	for i, p := range pkgs {
		for j, f := range p.GoFiles {
			if f == absPath {
				selected = true
				pkgIdx = i
				fileIdx = j
			}
		}
	}
	if !selected {
		return Analysis{}, fmt.Errorf("failed to found file %s", path)
	}

	return Analysis{
		fs,
		path,
		offset,
		pkgs,
		pkgIdx,
		fileIdx,
	}, nil
}

func (a *Analysis) makeFuncData(pkgName, filePath string, fd *ast.FuncDecl) FuncData {
	return FuncData{
		pkgName,
		filePath,
		fd.Name.String(),
		a.getOffset(fd.Pos()),
	}
}

func (a *Analysis) getOffset(pos token.Pos) int {
	return a.fs.Position(pos).Offset
}

func (a *Analysis) containCheck(n ast.Node) bool {
	beg := a.getOffset(n.Pos())
	end := a.getOffset(n.End())

	return beg <= a.offset && a.offset < end
}

func (a *Analysis) getCursorFunc(fn *ast.File) *ast.FuncDecl {
	for _, decl := range fn.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok && a.containCheck(fd) {
			return fd
		}
	}
	return nil
}

var pref = [...]string{"Test", "Benchmark", "Example"}

func (a *Analysis) getTestFunc(fn *ast.File, name, pre string) *ast.FuncDecl {
	for _, decl := range fn.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			if fd.Name.String() == pre+name || fd.Name.String() == pre+"_"+name {
				return fd
			}
		}
	}
	return nil
}

func objFuncCheck(obj types.Object) bool {
	if obj == nil {
		return false
	}
	_, ok := obj.(*types.Func)
	return ok
}

func (a *Analysis) getFuncObj(pkg *types.Package, name string) types.Object {
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

func (a *Analysis) GetFuncData() (interface{}, error) {

	pkg := a.pkgs[a.pkgIdx]
	fileNode := pkg.Syntax[a.fileIdx]

	fd := a.getCursorFunc(fileNode)
	if fd == nil {
		return nil, nil
	}
	funcName := fd.Name.String()
	ty := -1
	for i, pre := range pref {
		if strings.HasPrefix(funcName, pre) {
			ty = i
			funcName = strings.TrimPrefix(funcName, pre)
			break
		}
	}

	if funcName[0] == '_' {
		funcName = funcName[1:]
	}

	if ty != -1 {
		testFuncData := TestFuncData{a.makeFuncData(pkg.Name, pkg.GoFiles[a.fileIdx], fd), pref[ty]}
		obj := a.getFuncObj(pkg.Types, funcName)
		if obj == nil {
			return &TestFuncJson{"test", testFuncData, nil, nil}, nil
		}
		objPos := a.fs.Position(obj.Pos())
		objFuncData := FuncData{
			obj.Pkg().Name(),
			objPos.Filename,
			obj.Name(),
			objPos.Offset,
		}
		fp := TestFuncJson{"test", testFuncData, &objFuncData, nil} // TODO
		return fp, nil
	} else {
		mainFuncData := a.makeFuncData(pkg.Name, pkg.GoFiles[a.fileIdx], fd)
		for _, pre := range pref {
			for _, p := range a.pkgs {
				for j, f := range p.GoFiles {
					if strings.HasSuffix(f, "_test.go") {
						decl := a.getTestFunc(p.Syntax[j], fd.Name.String(), pre)
						if decl != nil {
							return MainFuncJson{"not test", mainFuncData, &TestFuncData{a.makeFuncData(p.Name, f, decl), pre}}, nil
						}
					}
				}
			}
		}
		return MainFuncJson{"not test", mainFuncData, nil}, nil
	}
}
