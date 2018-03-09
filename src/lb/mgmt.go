package lb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	User           = "admin"
	Password       = "admin"
	HTTPListenAddr = ":4444"
)

func (m *Manager) HttpServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok {
			if user != User || pass != Password {
				w.Header().Set("WWW-Authenticate", "Basic realm=\"LB\"")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		} else {
			// TODO is there a cleaner way to do this?
			w.Header().Set("WWW-Authenticate", "Basic realm=\"LB\"")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/":
			fmt.Fprintf(w, Index)
		case "/stats":
			fmt.Fprintf(w, m.Stats())
		case "/config":
			fmt.Fprintf(w, m.DumpConfig())
		default:
			fmt.Fprintf(w, "default response...\n")
		}
	})

	log.Fatal(http.ListenAndServe(HTTPListenAddr, nil))
}

func (m *Manager) DumpConfig() string {
	config, _ := LoadConfig(m.configFile, nil) // TODO handle error
	return dumpJSON(config)
}

func (m *Manager) Stats() string {
	/*s := Stat {
		Proxies: m.proxies,
	}*/
	b := make(map[string][]*Backend)
	for _, p := range m.proxies {
		//fmt.Println(p)
		b[p.Listen] = p.Backends
	}
	return dumpJSON(b)
}

// TODO handle error
func dumpJSON(i interface{}) string {
	r, err := json.Marshal(i)
	if err != nil {
		log.Println(err)
	}
	var out bytes.Buffer
	json.Indent(&out, r, "", "  ")
	return out.String()
}
