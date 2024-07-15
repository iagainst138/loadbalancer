package lb

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
)

type BackendConnection struct {
	Count   int
	Backend *Backend
}

type LeastConn struct {
	sync.Mutex
	Backends          map[string]*BackendConnection
	ActiveConnections map[string]*BackendConnection
	Min               int
	LastIndex         int
}

func NewLeastConn(backends []*Backend) *LeastConn {
	lc := LeastConn{
		Backends:          make(map[string]*BackendConnection),
		ActiveConnections: make(map[string]*BackendConnection),
	}

	for _, b := range backends {
		lc.Backends[b.Addr] = &BackendConnection{Backend: b}
	}

	return &lc
}

func (lc *LeastConn) HandleStarted(c net.Conn) {
}

func (lc *LeastConn) HandleDone(c net.Conn) {
	lc.Lock()
	lc.ActiveConnections[c.RemoteAddr().String()].Count -= 1
	delete(lc.ActiveConnections, c.RemoteAddr().String())
	lc.Unlock()
}

func (lc *LeastConn) NextBackend(c net.Conn) (*Backend, error) {

	var backend *Backend = nil

	keys := make([]string, 0, len(lc.Backends))
	for k := range lc.Backends {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, j := lc.LastIndex, 0; true; i++ {
		if i >= len(keys) {
			i = 0
		}
		backendConn := lc.Backends[keys[i]]
		if backendConn.Count <= lc.Min {
			backend = backendConn.Backend
			nextIndex := i + 1
			if nextIndex >= len(keys) {
				nextIndex = 0
			}
			lc.Lock()
			lc.LastIndex = nextIndex
			lc.Unlock()
			if backendConn.Count < lc.Min {
				// only break if its less than Min
				// else keep looping
				break
			}
		}
		j++
		if j >= len(keys) {
			break
		}
	}

	if backend == nil {
		return nil, errors.New("no backend found")
	}

	lc.Lock()
	defer lc.Unlock()
	lc.ActiveConnections[c.RemoteAddr().String()] = lc.Backends[backend.Addr]
	lc.ActiveConnections[c.RemoteAddr().String()].Count += 1
	lc.Min = lc.ActiveConnections[c.RemoteAddr().String()].Count
	return backend, nil
}

func (lc *LeastConn) Stats() string {
	return fmt.Sprintf("\nBackends: %v\nActiveConnections: %v\n", lc.Backends, len(lc.ActiveConnections))
}

func (lc *LeastConn) Name() string {
	return "LeastConn"
}
