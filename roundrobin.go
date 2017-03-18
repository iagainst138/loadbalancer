package main

import (
	"fmt"
	"net"
	"sync"
)

type RoundRobin struct {
	sync.Mutex
	backendIndex int
	Backends     []*Backend
}

func (r *RoundRobin) NextBackend(c net.Conn) (*Backend, error) {
	r.Lock()
	defer r.Unlock()
	r.backendIndex += 1
	if r.backendIndex > len(r.Backends)-1 {
		r.backendIndex = 0
	}
	return r.Backends[r.backendIndex], nil
}

func (r *RoundRobin) Stats() string {
	return fmt.Sprintf("\nBackends: %v\nIndex: %v\n", r.Backends, r.backendIndex)
}

func (r *RoundRobin) Name() string {
	return "RoundRobin"
}

func (r *RoundRobin) HandleStarted(c net.Conn) {
	// do nothing
}

func (r *RoundRobin) HandleDone(c net.Conn) {
	// do nothing
}
