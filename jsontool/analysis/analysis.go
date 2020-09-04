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
	FileName string
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
	TestFunc   []TestFuncData
}

// testにいる時に返したい値
type TestFuncJson struct {
	Type        string
	TestFunc    TestFuncData
	TargetFunc  FuncData
	SubTestName []string
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

func (a *Analysis) getTestFunc(fn *ast.File, name string) ([]*ast.FuncDecl, []int) {
	var fdList []*ast.FuncDecl
	var tyList []int
	for _, decl := range fn.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			for ty, pre := range pref {
				if fd.Name.String() == pre+name || fd.Name.String() == pre+"_"+name {
					fdList = append(fdList, fd)
					tyList = append(tyList, ty)
				}
			}
		}
	}
	return fdList, tyList
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

	// Test, Example, BenchMark
	if ty != -1 {
		obj := a.getFuncObj(pkg.Types, funcName)
		objPos := a.fs.Position(obj.Pos())
		objFuncData := FuncData{
			obj.Pkg().Name(),
			objPos.Filename,
			obj.Name(),
			objPos.Offset,
		}
		fp := TestFuncJson{"test", TestFuncData{a.makeFuncData(pkg.Name, pkg.GoFiles[a.fileIdx], fd), pref[ty]}, objFuncData, nil} // TODO
		return fp, nil
	} else {
		fp := MainFuncJson{"not test", a.makeFuncData(pkg.Name, pkg.GoFiles[a.fileIdx], fd), []TestFuncData{}}
		for _, p := range a.pkgs {
			for j, f := range p.GoFiles {
				if strings.HasSuffix(f, "_test.go") {
					declList, typeList := a.getTestFunc(p.Syntax[j], fd.Name.String())
					for k, _ := range declList {
						// TODO: 小文字
						fp.TestFunc = append(fp.TestFunc, TestFuncData{
							a.makeFuncData(p.Name, f, declList[k]),
							pref[typeList[k]],
						})
					}
				}
			}
		}
		return fp, nil
	}
}
