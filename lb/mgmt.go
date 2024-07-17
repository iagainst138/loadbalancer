package lb

// inspired by https://github.com/BlueDragonX/go-proxy-example

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
)

var (
	User           = "admin"
	Password       = "admin"
	HTTPListenAddr = ":4444"
)

type Proxies []*Proxy

func (p Proxies) Len() int      { return len(p) }
func (p Proxies) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p Proxies) Less(i, j int) bool {
	return p[i].Type+p[i].Listen < p[j].Type+p[j].Listen
}

type Manager struct {
	proxies        Proxies
	configFile     string
	doneChan       chan bool
	attemptingStop bool
	signalChan     chan os.Signal
}

func NewManager(configFile string) (*Manager, error) {
	doneChan := make(chan bool)
	m := Manager{
		configFile: configFile,
		doneChan:   doneChan,
		signalChan: make(chan os.Signal),
	}

	if err := m.reload(); err != nil {
		return nil, err
	}

	signal.Notify(m.signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go m.signalHandler()

	return &m, nil
}

func (m *Manager) reload() error {
	var errored bool

	config, err := LoadConfig(m.configFile, m.signalChan)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return err
	}
	m.proxies = nil
	proxies := make([]*Proxy, 0)
	for _, entry := range config.Entries {
		p, err := NewProxy(entry)
		if err != nil {
			slog.Error("call to NewProxy failed", "error", err, "entry", entry)
			errored = true
		} else {
			proxies = append(proxies, p)
		}
	}

	if errored {
		return errors.New("failed to load proxies")
	}

	m.proxies = proxies

	return nil
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
			log.Printf("attempting to stop: pid = %d", os.Getpid())
			if m.attemptingStop { // useful if connections enter a CLOSE_WAIT state
				log.Printf("forcing stop: pid = %d", os.Getpid())
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

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// set a custom header to identify the service
		w.Header().Set("Server", fmt.Sprintf("lb, version %v", Version))

		switch r.URL.Path {
		case "/":
			index, err := GetResource("resources/index.html")
			if err != nil {
				http.Error(w, "Error", http.StatusInternalServerError)
			} else {
				w.Write(index)
			}
		case "/stats":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, m.Stats())
		case "/config":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, m.DumpConfig())
		default:
			if strings.HasPrefix(r.URL.Path, "/static/") {
				ServeResource(w, r)
			} else {
				http.NotFound(w, r)
			}
		}
	})

	slog.Info("manager", "addr", HTTPListenAddr)
	log.Fatal(http.ListenAndServe(HTTPListenAddr, nil))
}

func (m *Manager) DumpConfig() string {
	config, _ := LoadConfig(m.configFile, nil) // TODO handle error
	return dumpJSON(config)
}

func (m *Manager) Stats() string {
	sort.Sort(m.proxies)
	return dumpJSON(m.proxies)
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
