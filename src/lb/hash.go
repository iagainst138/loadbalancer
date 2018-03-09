package lb

import (
	"fmt"
	"net"
)

type Hash struct {
	Backends []*Backend
}

// NOTE currently only works if all backends are up
func (h *Hash) NextBackend(c net.Conn) (*Backend, error) {
	// TODO could factor in the port also
	host, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		return nil, err
	}
	i := 0
	for _, b := range []byte(net.ParseIP(host)) {
		i += int(b)
	}
	return h.Backends[i%len(h.Backends)], nil
}

func (h *Hash) Name() string {
	return "Hash"
}

func (h *Hash) Stats() string {
	return fmt.Sprintf("\nBackends: %v\n", h.Backends)
}

func (h *Hash) HandleStarted(c net.Conn) {
	// do nothing
}

func (h *Hash) HandleDone(c net.Conn) {
	// do nothing
}
