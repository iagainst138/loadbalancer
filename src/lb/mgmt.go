package lb

// inspired by https://github.com/BlueDragonX/go-proxy-example

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	User           = "admin"
	Password       = "admin"
	HTTPListenAddr = ":4444"
)

type Manager struct {
	proxies        []*Proxy
	configFile     string
	doneChan       chan bool
	attemptingStop bool
	signalChan     chan os.Signal
}

func NewManager(configFile string) *Manager {
	doneChan := make(chan bool)
	m := Manager{
		configFile: configFile,
		doneChan:   doneChan,
		signalChan: make(chan os.Signal),
	}
	signal.Notify(m.signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go m.signalHandler()
	m.reload()
	return &m
}

func (m *Manager) reload() {
	config, err := LoadConfig(m.configFile, m.signalChan)
	if err != nil {
		log.Fatal(err)
	}
	m.proxies = nil
	proxies := make([]*Proxy, 0)
	for _, c := range config.Entries {
		p := NewProxy(c)
		proxies = append(proxies, p)
	}
	m.proxies = proxies
}

func (m *Manager) stopProxies() {
	for _, p := range m.proxies {
		log.Println("attempting to shut down:", p.Listen)
		if err := p.Close(); err != nil {
			log.Fatal(err.Error())
		} else {
			// TODO handle open connections
			for !p.Stopped {
				time.Sleep(2 * time.Millisecond)
			}
		}
	}
}

func (m *Manager) signalHandler() {
	for {
		receivedSignal := <-m.signalChan
		log.Println("received signal:", receivedSignal)
		if receivedSignal == syscall.SIGHUP {
			m.stopProxies()
			m.reload()
			m.Run()
		} else if receivedSignal == syscall.SIGTERM || receivedSignal == syscall.SIGINT {
			log.Println("attempting to stop: pid = ", os.Getpid())
			if m.attemptingStop { // useful if connections enter a CLOSE_WAIT state
				log.Println("forcing stop: pid = ", os.Getpid())
				os.Exit(1)
			} else {
				go func() {
					m.attemptingStop = true
					m.stopProxies()
					m.doneChan <- true
				}()
			}
		}
	}
}

func (m *Manager) Run() {
	for _, proxy := range m.proxies {
		go func(p *Proxy) {
			if err := p.Run(); err != nil {
				log.Println(err.Error())
			}
		}(proxy)
	}
}

func (m *Manager) Wait() bool {
	return <-m.doneChan
}

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
