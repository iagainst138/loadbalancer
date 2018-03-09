package main

// TODO line 239

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const BackoffTime = 500

type Balancer interface {
	NextBackend(net.Conn) (*Backend, error)
	HandleStarted(net.Conn)
	HandleDone(net.Conn)
	Stats() string
	Name() string
}

// Addr is an ip:port string
type Backend struct {
	Addr              string
	ActiveConnections int64
}

// TODO This should be able to dial TLS also
func (b *Backend) Dial(connType string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout(connType, b.Addr, timeout)
	return conn, err
}

func (b *Backend) DialUDP() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", b.Addr)
	if err != nil {
		return nil, err
	}
	backend, err := net.DialUDP("udp", nil, addr)
	return backend, err
}

func (b *Backend) inc() {
	atomic.AddInt64(&b.ActiveConnections, 1)
}

func (b *Backend) dec() {
	atomic.AddInt64(&b.ActiveConnections, -1)
}

// Proxy connections from Listen to Backend.
type Proxy struct {
	sync.Mutex
	listener net.Listener
	udpConn  *net.UDPConn
	Listen   string
	Type     string
	Backends []*Backend
	Balancer Balancer
	Stopped  bool
	Timeout  int
	CertFile string
	KeyFile  string
	useTls   bool
}

func NewProxy(entry *Entry) *Proxy {
	proxy := Proxy{
		Listen:   entry.ListenAddr,
		Backends: entry.Backends,
		Type:     entry.Type,
		Timeout:  entry.Timeout,
	}

	// UDP only supports RoundRobin
	if entry.Type == "udp" {
		proxy.Balancer = &RoundRobin{Backends: entry.Backends}
	} else {
		switch entry.Backend {
		case "RoundRobin":
			proxy.Balancer = &RoundRobin{Backends: entry.Backends}
		case "Hash":
			proxy.Balancer = &Hash{Backends: entry.Backends}
		case "LeastConn":
			proxy.Balancer = NewLeastConn(proxy.Backends)
		default:
			log.Fatal("error: upsupported backend '%s'")
		}
	}

	if entry.CertFile != "" && entry.KeyFile != "" {
		proxy.useTls = true
		proxy.CertFile = entry.CertFile
		proxy.KeyFile = entry.KeyFile
	}
	return &proxy
}

// TODO improve this output
func (p *Proxy) Stats() {
	logGreen(p.Listen)
	log.Printf("%v [%v]", p.Balancer.Name(), p.Type)
	log.Println(p.Balancer.Stats())
}

func (p *Proxy) listenUDP() error {
	listen_addr, err := net.ResolveUDPAddr("udp", p.Listen)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", listen_addr)
	if err != nil {
		log.Fatal(err)
	}
	p.udpConn = conn

	log.Printf("[udp] listening on %v\n", p.Listen)

	for {
		buffer := make([]byte, 4096)
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Println(err)
			break
		}

		backend, err := p.Balancer.NextBackend(conn)

		go func(bytes_read int, client_addr net.Addr) {
			backend_conn, err := backend.DialUDP()
			if err != nil {
				log.Printf("error: %v\n", err)
				return
			}
			defer backend_conn.Close()
			_, err = backend_conn.Write(buffer[:bytes_read]) // TODO validate the number of bytes written
			if err != nil {
				log.Printf("error: %v\n", err)
				return
			}
			b := make([]byte, 4096)
			r, _, err := backend_conn.ReadFrom(b)
			if err != nil {
				log.Println(err)
				return
			}
			_, err = conn.WriteTo(b[:r], client_addr) // TODO validate the number of bytes written
			if err != nil {
				log.Printf("error: %v\n", err)
			}
		}(n, addr)
	}

	return nil
}

func (p *Proxy) listenTCP() error {
	var err error

	if p.useTls {
		cert, err := tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
		if err != nil {
			log.Fatalf("server: loadkeys: %s", err)
		}
		config := tls.Config{Certificates: []tls.Certificate{cert}}
		p.listener, err = tls.Listen(p.Type, p.Listen, &config)
		if err != nil {
			log.Fatalf("server: listen: %s", err)
		}
	} else if p.listener, err = net.Listen(p.Type, p.Listen); err != nil {
		return err
	}

	tlsMessage := ""
	if p.useTls {
		tlsMessage = "[TLS " + p.CertFile + " " + p.KeyFile + "]"
	}
	log.Printf("[tcp] listening on %s %s", p.Listen, tlsMessage)

	errorMessage := ""
	wg := &sync.WaitGroup{}
	for {
		if conn, err := p.listener.Accept(); err == nil {
			wg.Add(1)
			go func(nc net.Conn) {
				defer wg.Done()
				p.handleTCP(nc)
			}(conn)
		} else {
			// NOTE this is reached when Close() is called on net.Listener
			// NOTE "too many open files" can be reached here
			log.Println("listen failed:", err)
			if strings.Index(err.Error(), "too many open files") == -1 {
				errorMessage = err.Error()
				break
			} else {
				logRed("backing off: " + err.Error())
				time.Sleep(BackoffTime * time.Millisecond)
			}
		}
	}
	wg.Wait()
	log.Printf("proxy %s stopped", p.Listen)
	log.Println(errorMessage)
	p.Stopped = true
	return nil
}

func (p *Proxy) Run() error {
	if p.Type == "udp" {
		return p.listenUDP()
	} else if p.Type == "tcp" {
		return p.listenTCP()
	}
	return errors.New("unknown type: " + p.Type)
}

func (p *Proxy) Close() error {
	if p.listener != nil {
		return p.listener.Close()
	} else if p.udpConn != nil {
		return p.udpConn.Close()
	}
	return nil
}

func (p *Proxy) handleTCP(conn net.Conn) {
	defer conn.Close()
	p.Balancer.HandleStarted(conn)

	blacklist := make(map[string]int)

	for attempts := 0; attempts < len(p.Backends); attempts++ {
		backend, err := p.Balancer.NextBackend(conn)
		if err != nil {
			log.Printf("error getting backend: %s", err)
			return
		}
		if _, exists := blacklist[backend.Addr]; exists {
			blacklist[backend.Addr]++
			if blacklist[backend.Addr] > len(p.Backends) {
				log.Println("*** breaking due to blacklist ***") // TODO review
				break
			}
			attempts-- // attempt to hit all backends
			continue
		}
		backendConn, err := backend.Dial(p.Type, time.Duration(p.Timeout)*time.Second)
		if err != nil {
			log.Println(err)
			blacklist[backend.Addr] = 0
		} else {
			backend.inc()
			defer backendConn.Close()
			defer p.Balancer.HandleDone(conn)
			defer backend.dec()
			if cError, bError := p.Pipe(conn, backendConn); cError != nil || bError != nil {
				log.Printf("pipe failed:\n%v\n%v\n", cError, bError)
			}
			return // exit the attempt loop
		}
	}
	logRed("failed to reach a running backend")
}

func (p *Proxy) Pipe(client, backend net.Conn) (error, error) {
	defer client.Close()
	defer backend.Close()

	var clientCopyError error
	var backendCopyError error
	var clientCopied int64
	var backendCopied int64
	clientOK := false

	var wg sync.WaitGroup
	wg.Add(2)

	// copy client data to backend
	go func() {
		defer backend.Close()
		defer wg.Done()
		if clientCopied, clientCopyError = io.Copy(backend, client); clientCopyError == nil {
			// client has closed connection so we mark it as ok
			clientOK = true
		}
	}()

	// copy backend data to client
	go func() {
		defer client.Close()
		defer wg.Done()
		backendCopied, backendCopyError = io.Copy(client, backend)
		if clientOK {
			backendCopyError = nil
		}
	}()

	wg.Wait()
	return clientCopyError, backendCopyError
}
