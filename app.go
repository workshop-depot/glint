package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/urfave/cli"
)

func cmdHelpers(*cli.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	_forAllDirs(buf, wd)
	fmt.Println(string(buf.Bytes()))

	return nil
}

func _forAllDirs(buf *bytes.Buffer, wd string) {
	err := filepath.Walk(wd, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !f.IsDir() {
			return nil
		}
		if strings.HasPrefix(f.Name(), ".") {
			return filepath.SkipDir
		}
		_forDir(buf, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func _forDir(buf *bytes.Buffer, wd string) {
	pkgs, err := _getPackages(wd)
	if err != nil {
		panic(err)
	}

	funcs := _extractFuncs(pkgs)
	helperFunctionNames := _extractHelperNames(funcs)
	sourceFiles, err := _listsourceFiles(wd)
	if err != nil {
		panic(err)
	}

	_writePkgInfo(wd, buf)
	_study(buf, sourceFiles, helperFunctionNames)
}

func _study(buf *bytes.Buffer, sourceFiles, helperFunctionNames []string) {
	anyErrors := false
	rn := regexp.MustCompile("_\\d+")
	for _, vsrc := range sourceFiles {
		b, err := ioutil.ReadFile(vsrc)
		if err != nil {
			panic(err)
		}

		anyErrors = _studyCalls(buf, rn, string(b), helperFunctionNames) || anyErrors
	}
	if !anyErrors {
		fmt.Fprintln(buf, "OK")
	}
}

func _studyCalls(
	buf *bytes.Buffer,
	rgx *regexp.Regexp,
	content string,
	helperFunctionNames []string) (anyErrors bool) {
	for _, vfn := range helperFunctionNames {
		hrgx := regexp.MustCompile("(?P<func_name>" + vfn + ")\\W")
		matches := hrgx.FindStringSubmatch(content)
		var all []string
		strmatches := make(map[string]string)
		for k, v := range hrgx.SubexpNames() {
			if v == "" || len(matches) <= k {
				continue
			}
			strmatches[v] = matches[k]
		}
		for _, v := range strmatches {
			for i := 0; i < strings.Count(content, v); i++ {
				all = append(all, v)
			}
		}

		expected, actual := _listCalls(rgx, all)

		for k, v := range actual {
			v := v - 1
			if v == expected[k] {
				continue
			}
			fmt.Fprintf(buf, "% -18v actual calls: % 3d expected calls: % 3d \n", k+"(...)", v, expected[k])
			anyErrors = true
		}
	}
	return
}

func _listCalls(rgx *regexp.Regexp, all []string) (expected, actual map[string]int) {
	expected = make(map[string]int)
	actual = make(map[string]int)
	for _, v := range all {
		v := strings.TrimSpace(v)
		actual[v] = actual[v] + 1
		_, ok := expected[v]
		if ok {
			continue
		}
		numbers := rgx.FindAllString(v, -1)
		var nstr = "1"
		if len(numbers) > 0 {
			nstr = strings.Replace(numbers[0], "_", "", -1)
		}
		n, err := strconv.Atoi(nstr)
		if err != nil {
			panic(err)
		}
		expected[v] = n
	}
	return
}

func _writePkgInfo(wd string, buf *bytes.Buffer) {
	s := strings.Replace(wd, build.Default.GOPATH, "", -1)
	s = strings.TrimPrefix(s, "/src/")
	fmt.Fprintf(buf, "package: %s\n", s)
}

func _getPackages(dir string) (pkgs map[string]*ast.Package, err error) {
	set := token.NewFileSet()
	pkgs, err = parser.ParseDir(set, dir, nil, 0)
	if err != nil {
		return nil, err
	}
	return
}

func _extractFuncs(pkgs map[string]*ast.Package) (funcs []*ast.FuncDecl) {
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, d := range f.Decls {
				if fn, isFn := d.(*ast.FuncDecl); isFn {
					funcs = append(funcs, fn)
				}
			}
		}
	}
	return
}

func _extractHelperNames(funcs []*ast.FuncDecl) (fhp []string) {
	for _, vf := range funcs {
		fn := fmt.Sprint(vf.Name)
		if !strings.HasPrefix(fn, "_") {
			continue
		}
		fhp = append(fhp, fn)
	}
	return
}

func _listsourceFiles(wd string) (sourceFiles []string, err error) {
	files, err := ioutil.ReadDir(wd)
	if err != nil {
		return nil, err
	}
	for _, vff := range files {
		fn := vff.Name()
		if filepath.HasPrefix(fn, ".") {
			continue
		}
		if strings.ToLower(filepath.Ext(fn)) != ".go" {
			continue
		}
		sourceFiles = append(sourceFiles, filepath.Join(wd, fn))
	}
	return
}
