package DHT

import (
	"errors"
	"net"
	"strconv"
	"syscall"
)

func ignoreReadFromError(err error) bool {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return false
	}
	//return errors.Is(errno, syscall.WSAECONNRESET)
	return true
}

func AddrPort(addr net.Addr) int {
	switch raw := addr.(type) {
	case *net.UDPAddr:
		return raw.Port
	case *net.TCPAddr:
		return raw.Port
	default:
		_, port, err := net.SplitHostPort(addr.String())
		if err != nil {
			panic(err)
		}
		i64, err := strconv.ParseInt(port, 0, 0)
		if err != nil {
			panic(err)
		}
		return int(i64)
	}
}

func AddrIP(addr net.Addr) net.IP {
	if addr == nil {
		return nil
	}
	switch raw := addr.(type) {
	case *net.UDPAddr:
		return raw.IP
	case *net.TCPAddr:
		return raw.IP
	default:
		host, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			panic(err)
		}
		return net.ParseIP(host)
	}
}

func inList[T uint64 | string](l []T, e T) bool {
	for _, t := range l {
		if t == e {
			return true
		}
	}
	return false
}
