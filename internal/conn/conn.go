package conn

import (
	"errors"
	"io"
	"strings"
	"syscall"
)

type Conn interface {
	io.ReadWriteCloser

	Name() string

	Info() string

	Dial(dst string) (Conn, error)

	Listen(dst string) (Conn, error)
	Accept() (Conn, error)
}

func NewConn(proto string) (Conn, error) {
	proto = strings.ToLower(proto)
	switch proto {
	case "tcp":
		return &TcpConn{}, nil
	case "quic":
		return &QuicConn{}, nil
	}
	return nil, errors.New("undefined proto " + proto)
}

func SupportReliableProtos() []string {
	ret := make([]string, 0, 2)
	ret = append(ret, "tcp")
	ret = append(ret, "quic")
	return ret
}

func SupportProtos() []string {
	ret := make([]string, 0)
	ret = append(ret, SupportReliableProtos()...)
	ret = append(ret, "udp")
	return ret
}

func HasReliableProto(proto string) bool {
	return hasString(SupportReliableProtos(), proto)
}

func HasProto(proto string) bool {
	return hasString(SupportProtos(), proto)
}

func hasString(data []string, dst string) bool {
	for _, i := range data {
		if i == dst {
			return true
		}
	}
	return false
}

var gControlOnConnSetup func(network, address string, c syscall.RawConn) error

func RegisterDialerController(fn func(network, address string, c syscall.RawConn) error) {
	gControlOnConnSetup = fn
}
