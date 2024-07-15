package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

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
				log.Printf("adding '%v' to resources\n", path)
				b, err := ioutil.ReadFile(path)
				if err != nil {
					log.Fatalf("error: %s", err)
				}
				files[strings.Replace(path, "\\", "/", -1)] = base64.StdEncoding.EncodeToString(b)
			}
			return nil
		})
	}
	return files, nil
}

func main() {
	resourceFiles := "resources,static"

	flag.StringVar(&resourceFiles, "resources", resourceFiles, "comma separated list of files to add")
	flag.Parse()

	files, err := GenFileList(resourceFiles)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	v := Vars{
		Files:   files,
		DevMode: os.Getenv("LB_DEV_MODE") == "1",
	}

	t, err := template.ParseFiles("lb/templates/resources")
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	f, err := os.Create("lb/resources.go")
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	if err := (t.Execute(f, v)); err != nil {
		log.Fatalf("error: %s", err)
	}
}
