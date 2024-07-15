package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func check(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

// Vars holds variables needed for the resources template
type Vars struct {
	Files   map[string]string
	DevMode bool
}

func GenFileList(file string) (map[string]string, error) {
	files := make(map[string]string)
	for _, f := range strings.Split(file, ",") {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return nil, fmt.Errorf("error: %v does not exist\n", f)
		}
		filepath.Walk(f, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				fmt.Printf("adding '%v' to resources\n", path)
				b, e := ioutil.ReadFile(path)
				check(e)
				files[strings.Replace(path, "\\", "/", -1)] = base64.StdEncoding.EncodeToString(b)
			}
			return nil
		})
	}
	return files, nil
}

func main() {
	dir, _ := os.Getwd()
	fmt.Println("working dir:", dir)

	resourceFiles := "resources,static"

	flag.StringVar(&resourceFiles, "resources", resourceFiles, "comma separated list of files to add")
	flag.Parse()

	files, err := GenFileList(resourceFiles)
	check(err)

	v := Vars{
		Files:   files,
		DevMode: os.Getenv("LB_DEV_MODE") == "1",
	}

	t, err := template.ParseFiles("src/lb/templates/resources")
	check(err)

	f, err := os.Create("src/lb/resources.go")
	check(err)
	check(t.Execute(f, v))
}
