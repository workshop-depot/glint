package main

import (
	"bytes"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/urfave/cli"
)

var conf struct{}

func cmdHelpers(*cli.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pkgList := make(map[string][]string)

	filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		rel, err := filepath.Rel(wd, path)
		if err != nil {
			panic(err)
		}
		if filepath.HasPrefix(rel, ".") {
			return nil
		}
		if strings.ToLower(filepath.Ext(rel)) != ".go" {
			return nil
		}

		pkgList[filepath.Dir(rel)] = append(pkgList[filepath.Dir(rel)], path)

		return nil
	})

	r := regexp.MustCompile("\\s+_\\d*\\S+\\(")
	rn := regexp.MustCompile("_\\d+")
	anyErrors := false
	for rel, files := range pkgList {
		_ = rel
		buf := &bytes.Buffer{}
		if len(files) > 0 {
			s := strings.Replace(files[0], build.Default.GOPATH, "", -1)
			s = filepath.Dir(s)
			s = strings.TrimPrefix(s, "/src/")
			fmt.Fprintf(buf, "package: %s\n", s)
		}
		for _, path := range files {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			all := r.FindAllString(string(b), -1)

			expected := make(map[string]int)
			actual := make(map[string]int)
			for _, v := range all {
				v := strings.TrimSpace(v)
				actual[v] = actual[v] + 1
				_, ok := expected[v]
				if ok {
					continue
				}
				numbers := rn.FindAllString(v, -1)
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

			for k, v := range actual {
				v := v - 1
				if v == expected[k] {
					continue
				}
				fmt.Fprintf(buf, "% -18v actual calls: % 3d expected calls: % 3d \n", k+"...)", v, expected[k])
				anyErrors = true
			}
		}
		if !anyErrors {
			fmt.Fprintln(buf, "OK")
		}
		fmt.Printf("%s", buf.Bytes())
	}

	return nil
}
