package lrp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"

	"github.com/sirupsen/logrus"
)

func NewLrpc(log *logrus.Logger, server, localAddr, proxy, notify string) *Lrpc {
	return &Lrpc{log, nil, server, localAddr, proxy, notify}
}

func NewLrps(log *logrus.Logger, pp, np string, addFunc func(*ProxyClient) bool) *Lrps {
	return &Lrps{log, pp, np, addFunc}
}

func NewProxyClient(conn net.Conn, siteId string) *ProxyClient {
	return &ProxyClient{conn, false, false, "", siteId}
}

type ProxyClient struct {
	conn   net.Conn
	lock   bool
	IsOpen bool
	Port   string
	SiteId string
}

func (pc *ProxyClient) Open() error {
	if pc.lock {
		return errors.New("wait other options")
	}
	pc.lock = true
	defer func() { pc.lock = false }()

	if _, err := pc.conn.Write([]byte{1}); err != nil {
		return err
	} else {
		status := make([]byte, 1)
		if _, err := pc.conn.Read(status); err != nil {
			return err
		}
		if status[0] == 0 {
			return errors.New("open proxy fail , check client log")
		} else {
			res := make([]byte, 8)
			if _, err := io.ReadFull(pc.conn, res); err != nil {
				return err
			} else {
				pc.Port = ":" + strconv.Itoa(pc.b2i(res))
				pc.IsOpen = true
				return nil
			}
		}
	}
}

func (pc *ProxyClient) Close() error {
	if pc.lock {
		return errors.New("wait other options")
	}
	pc.lock = true
	defer func() { pc.lock = false }()

	if _, err := pc.conn.Write([]byte{0}); err != nil {
		return err
	} else {
		status := make([]byte, 1)
		if _, err := pc.conn.Read(status); err != nil {
			return err
		}
		if status[0] == 0 {
			return errors.New("close proxy fail , check client log")
		} else {
			pc.Port = ""
			pc.IsOpen = false
			return nil
		}
	}
}

func (pc *ProxyClient) Destroy() {
	pc.conn.Close()
}

func (pc *ProxyClient) b2i(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}
