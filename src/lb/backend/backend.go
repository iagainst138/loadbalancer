package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

type Backend struct {
	Addr  string
	Port  string
	Sleep bool
}

func (b *Backend) Listen() {
	listen := net.JoinHostPort(b.Addr, b.Port)
	log.Println("[HTTP] listening on", listen)
	log.Fatal(http.ListenAndServe(listen, b))
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sleep := 0
	if b.Sleep && rand.Intn(20) > 15 {
		sleep = rand.Intn(10)
		if sleep == 1 {
			log.Printf("%v: dropping connection", b.Port)
			return
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	w.Write([]byte(fmt.Sprintf("port: %v (sleep %v)\n", b.Port, sleep)))
}

func main() {
	rand.Seed(time.Now().UnixNano())

	sleep := false
	addr := "127.0.0.1"

	flag.BoolVar(&sleep, "sleep", sleep, "enable random sleep")
	flag.StringVar(&addr, "listen-addr", addr, "address to listen on")
	flag.Parse()

	backendPorts := []string{
		"7000",
		"7001",
		"7002",
		"7003",
		"7004",
		"7005",
		"7006",
		"7007",
		"7008",
		"7009",
		"7010",
	}

	var wg sync.WaitGroup

	for _, port := range backendPorts {
		go func(p string) {
			wg.Add(1)
			b := Backend{Addr: addr, Port: p, Sleep: sleep}
			b.Listen()
			wg.Done()
		}(port)
	}

	wg.Wait()
}
