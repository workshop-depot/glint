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
	wd := _getwd()

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

		data := _fetchData(path)
		for pkg, pkgData := range data {
			_studyNames(buf, pkg, pkgData)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}

func _studyNames(buf *bytes.Buffer, pkg string, pkgData map[string][]string) {
	helperCalls := _helperCalls(pkgData)
	rgx := regexp.MustCompile("_\\d+")
	pkgErr := make(map[string][]string)
	for k, v := range helperCalls {
		pkgErr[v.pkg] = append(pkgErr[v.pkg])
		numbers := rgx.FindAllString(k, -1)
		var nstr = "1"
		if len(numbers) > 0 {
			nstr = strings.Replace(numbers[0], "_", "", -1)
		}
		n, err := strconv.Atoi(nstr)
		if err != nil {
			panic(err)
		}
		actual := v.count - 1
		expected := n
		if expected == actual {
			continue
		}
		errstr := fmt.Sprintf("% -18v actual calls: % 3d expected calls: % 3d", k+"(...)", actual, expected)
		pkgErr[v.pkg] = append(pkgErr[v.pkg], errstr)
	}
	for k, v := range pkgErr {
		if len(v) == 0 {
			fmt.Fprintf(buf, "%v %v\n", "OK", k)
			continue
		}
		fmt.Fprintf(buf, "%v:\n", k)
		for _, verr := range v {
			fmt.Fprintf(buf, "%v\n", verr)
		}
	}
}

func _helperCalls(pkgData map[string][]string) map[string]struct {
	count int
	pkg   string
} {
	helperCalls := make(map[string]struct {
		count int
		pkg   string
	})

	hm := make(map[string]struct{})
	fm := make(map[string]struct{})
	for file, helpers := range pkgData {
		fm[file] = struct{}{}
		for _, fn := range helpers {
			hm[fn] = struct{}{}
		}
	}

	var helpers []string
	for k := range hm {
		helpers = append(helpers, k)
	}
	var files []string
	for k := range fm {
		files = append(files, k)
	}

	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}
		content := string(b)
		for _, fn := range helpers {
			hrgx := regexp.MustCompile("\\W(?P<" + funcGroupName + ">" + fn + ")\\W")
			matches := hrgx.FindAllStringSubmatch(content, -1)

			for _, v := range hrgx.SubexpNames() {
				if v != funcGroupName {
					continue
				}
				count := 0
				for _, fcset := range matches {
					for _, fc := range fcset {
						if fc == fn {
							count++
						}
					}
				}

				fitem := helperCalls[fn]
				fitem.count += count
				fitem.pkg = _pkgName(file)
				helperCalls[fn] = fitem
			}
		}
	}
	return helperCalls
}

func _pkgName(d string) string {
	s := strings.Replace(d, build.Default.GOPATH, "", -1)
	s = strings.TrimPrefix(s, "/src/")
	s = filepath.Dir(s)
	return s
}

const funcGroupName = "func_name"

func _fetchData(currentDir string) map[string]map[string][]string {
	data := make(map[string]map[string][]string)
	pkgSet := _fetchPackages(currentDir)
	for _, vp := range pkgSet {
		for kf, vf := range vp.Files {
			for _, vd := range vf.Decls {
				if fn, isFn := vd.(*ast.FuncDecl); isFn {
					fname := fmt.Sprint(fn.Name)

					if data[vp.Name] == nil {
						data[vp.Name] = make(map[string][]string)
					}
					data[vp.Name][kf] = append(data[vp.Name][kf])

					if !strings.HasPrefix(fname, "_") {
						continue
					}
					data[vp.Name][kf] = append(data[vp.Name][kf], fname)
				}
			}
		}
	}
	return data
}

func _fetchPackages(dir string) (pkgs map[string]*ast.Package) {
	set := token.NewFileSet()
	var err error
	pkgs, err = parser.ParseDir(set, dir, nil, 0)
	if err != nil {
		panic(err)
	}
	return
}

func _getwd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}
