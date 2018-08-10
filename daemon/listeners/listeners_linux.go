package listeners // import "github.com/moooofly/hunter-agent/daemon/listeners"

import (
	"fmt"
	"net"

	"github.com/docker/go-connections/sockets"
)

// Init creates new listeners for the server.
func Init(proto, addr string) ([]net.Listener, error) {
	ls := []net.Listener{}

	switch proto {
	case "tcp":
		l, err := sockets.NewTCPSocket(addr, nil)
		if err != nil {
			return nil, err
		}
		ls = append(ls, l)
	case "unix":
		l, err := sockets.NewUnixSocket(addr, 0)
		if err != nil {
			return nil, fmt.Errorf("can't create unix socket %s: %v", addr, err)
		}
		ls = append(ls, l)
	default:
		return nil, fmt.Errorf("invalid protocol format: %q", proto)
	}

	return ls, nil
}
