package opts // import "github.com/moooofly/hunter-agent/opts"

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// DefaultUnixSocket Path for the unix socket.
// hunter-agent by default always listens on the default unix socket.
var DefaultUnixSocket = "/var/run/hunter-agent.sock"

// DefaultTCPPort Default TCP Port used if only the protocol is provided to -H flag e.g. hunter-agent -H tcp://
var DefaultTCPPort = 12345 // Default TCP Port
// DefaultTCPHost Default TCP Host used if only port is provided to -H flag e.g. hunter-agent -H tcp://:8080
const DefaultTCPHost = "0.0.0.0"

// DefaultTCPEndpoint constant defines the default host string used by hunter-agent.
var DefaultTCPEndpoint = fmt.Sprintf("tcp://%s:%d", DefaultTCPHost, DefaultTCPPort)

// DefaultUnixEndpoint constant defines the default host string used by hunter-agent and exporter
var DefaultUnixEndpoint = fmt.Sprintf("unix://%s", DefaultUnixSocket)

// DefaultBrokerHost defines the default broker host string used by hunter-agent.
var DefaultBrokerHost = "tcp://0.0.0.0:9092"

// ValidateBrokerHost validates that the specified string is a valid broker host and returns it.
func ValidateBrokerHost(val string) (string, error) {
	host := strings.TrimSpace(val)
	// The empty string means default and is not handled by parseDaemonHost
	if host != "" {
		_, err := parseBrokerHost(host)
		if err != nil {
			return val, err
		}
	}
	return val, nil
}

// ParseBrokerHost set defaults for a Daemon host string
func ParseBrokerHost(val string) (string, error) {
	host := strings.TrimSpace(val)
	if host == "" {
		host = DefaultBrokerHost
	} else {
		var err error
		host, err = parseBrokerHost(host)
		if err != nil {
			return val, err
		}
	}
	return host, nil
}

// parseBrokerHost parses the specified address and returns an address that will be used as broker host.
func parseBrokerHost(addr string) (string, error) {
	addrParts := strings.SplitN(addr, "://", 2)
	if len(addrParts) == 1 && addrParts[0] != "" {
		addrParts = []string{"tcp", addrParts[0]}
	}

	switch addrParts[0] {
	case "tcp":
		return ParseTCPAddr(addrParts[1], DefaultBrokerHost)
	default:
		return "", fmt.Errorf("Invalid broker address format: %s", addr)
	}
}

// ValidateHost validates that the specified string is a valid host and returns it.
func ValidateHost(val string) (string, error) {
	host := strings.TrimSpace(val)
	// The empty string means default and is not handled by parseDaemonHost
	if host != "" {
		_, err := parseDaemonHost(host)
		if err != nil {
			return val, err
		}
	}
	return val, nil
}

// ParseHost and set defaults for a Daemon host string
func ParseHost(val string) (string, error) {
	host := strings.TrimSpace(val)
	if host == "" {
		host = DefaultUnixEndpoint
	} else {
		var err error
		host, err = parseDaemonHost(host)
		if err != nil {
			return val, err
		}
	}
	return host, nil
}

// parseDaemonHost parses the specified address and returns an address that will be used as the host.
// Depending of the address specified, this may return one of the global Default* strings defined in hosts.go.
func parseDaemonHost(addr string) (string, error) {
	addrParts := strings.SplitN(addr, "://", 2)
	if len(addrParts) == 1 && addrParts[0] != "" {
		addrParts = []string{"tcp", addrParts[0]}
	}

	switch addrParts[0] {
	case "tcp":
		return ParseTCPAddr(addrParts[1], DefaultTCPEndpoint)
	case "unix":
		return parseSimpleProtoAddr("unix", addrParts[1], DefaultUnixSocket)
	default:
		return "", fmt.Errorf("Invalid bind address format: %s", addr)
	}
}

// parseSimpleProtoAddr parses and validates that the specified address is a valid
// socket address for simple protocols like unix and npipe. It returns a formatted
// socket address, either using the address parsed from addr, or the contents of
// defaultAddr if addr is a blank string.
func parseSimpleProtoAddr(proto, addr, defaultAddr string) (string, error) {
	addr = strings.TrimPrefix(addr, proto+"://")
	if strings.Contains(addr, "://") {
		return "", fmt.Errorf("Invalid proto, expected %s: %s", proto, addr)
	}
	if addr == "" {
		addr = defaultAddr
	}
	return fmt.Sprintf("%s://%s", proto, addr), nil
}

// ParseTCPAddr parses and validates that the specified address is a valid TCP
// address. It returns a formatted TCP address, either using the address parsed
// from tryAddr, or the contents of defaultAddr if tryAddr is a blank string.
// tryAddr is expected to have already been Trim()'d
// defaultAddr must be in the full `tcp://host:port` form
func ParseTCPAddr(tryAddr string, defaultAddr string) (string, error) {
	if tryAddr == "" || tryAddr == "tcp://" {
		return defaultAddr, nil
	}
	addr := strings.TrimPrefix(tryAddr, "tcp://")
	if strings.Contains(addr, "://") || addr == "" {
		return "", fmt.Errorf("Invalid proto, expected tcp: %s", tryAddr)
	}

	defaultAddr = strings.TrimPrefix(defaultAddr, "tcp://")
	defaultHost, defaultPort, err := net.SplitHostPort(defaultAddr)
	if err != nil {
		return "", err
	}
	// url.Parse fails for trailing colon on IPv6 brackets on Go 1.5, but
	// not 1.4. See https://github.com/golang/go/issues/12200 and
	// https://github.com/golang/go/issues/6530.
	if strings.HasSuffix(addr, "]:") {
		addr += defaultPort
	}

	u, err := url.Parse("tcp://" + addr)
	if err != nil {
		return "", err
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		// try port addition once
		host, port, err = net.SplitHostPort(net.JoinHostPort(u.Host, defaultPort))
	}
	if err != nil {
		return "", fmt.Errorf("Invalid bind address format: %s", tryAddr)
	}

	if host == "" {
		host = defaultHost
	}
	if port == "" {
		port = defaultPort
	}
	p, err := strconv.Atoi(port)
	if err != nil && p == 0 {
		return "", fmt.Errorf("Invalid bind address format: %s", tryAddr)
	}

	return fmt.Sprintf("tcp://%s%s", net.JoinHostPort(host, port), u.Path), nil
}
