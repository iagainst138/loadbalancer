package main

// inspired by https://github.com/BlueDragonX/go-proxy-example

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Manager struct {
	proxies        []*Proxy
	configFile     string
	doneChan       chan bool
	attemptingStop bool
}

func NewManager(configFile string) *Manager {
	doneChan := make(chan bool)
	m := Manager{
		configFile: configFile,
		doneChan:   doneChan,
	}
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGWINCH)
	go m.signalHandler(signalChan)
	m.reload()
	return &m
}

func (m *Manager) reload() {
	config, err := LoadConfig(m.configFile)
	if err != nil {
		log.Fatal(err)
	}
	m.proxies = nil
	proxies := make([]*Proxy, 0)
	for _, c := range config {
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

func (m *Manager) signalHandler(signalChan chan os.Signal) {
	for {
		receivedSignal := <-signalChan
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
		} else if receivedSignal == syscall.SIGWINCH {
			m.Stats()
		}
	}
}

func (m *Manager) Stats() {
	for _, p := range m.proxies {
		p.Stats()
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

func main() {
	configFile := ""
	pidFile := ""

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&configFile, "config", "", "config")
	flags.StringVar(&pidFile, "pid-file", "", "write pid to this file")
	flags.Parse(os.Args[1:])

	if configFile == "" {
		log.Fatal("error: no config specified")
	}

	if pidFile != "" {
		if err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
			log.Fatal(err)
		}
		log.Printf("pid [%v] written to %v", os.Getpid(), pidFile)
	}

	m := NewManager(configFile)
	m.Run()
	m.Wait()
	log.Println("terminating")
}