package lb

import (
	"embed"
	"net/http"
	"os"
	"strings"
)

//go:embed resources/*
var fs embed.FS

var DevMode = false

func GetResource(key string) ([]byte, error) {
	if DevMode {
		return os.ReadFile(key)
	} else {
		return fs.ReadFile(key)
	}
}

func ServeResource(w http.ResponseWriter, req *http.Request) {
	key := strings.Replace(req.URL.Path, "/", "", 1)
	b, err := GetResource(key)
	if err != nil {
		http.NotFound(w, req)
	} else {
		if strings.HasSuffix(key, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		// TODO set better headers
		w.Write(b)
	}
}
