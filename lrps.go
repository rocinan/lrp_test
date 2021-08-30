package lrp

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"unsafe"

	"github.com/lev2048/lrp/internal/server"

	"github.com/sirupsen/logrus"
)

type Lrps struct {
	log        *logrus.Logger
	proxyPort  string
	notifyPort string
	addProxy   func(*ProxyClient) bool
}

func (ls *Lrps) Run() bool {
	server := server.NewLrpServer(ls.log, ls.proxyPort, context.Background())
	if err := server.Run(); err != nil {
		ls.log.WithField("info", err).Warn("proxy port listen error", err)
		return false
	}

	if ln, err := net.Listen("tcp4", ls.notifyPort); err != nil {
		ls.log.Warn("notify port listen error: ", err)
		return false
	} else {
		go func() {
			for {
				if conn, err := ln.Accept(); err != nil {
					ls.log.Warn("accept error:", err)
					break
				} else {
					go ls.handleConn(conn)
				}
			}
		}()
	}
	return true
}

func (ls *Lrps) handleConn(conn net.Conn) {
	len := make([]byte, 8)
	if n, err := conn.Read(len); n != 8 || err != nil {
		ls.log.Warn("read error:", err)
		conn.Close()
		return
	} else {
		buf := make([]byte, ls.b2i(len))
		if _, err := io.ReadFull(conn, buf); err != nil {
			ls.log.Warn("read siteId error:", err)
			conn.Close()
		} else {
			ls.log.Info("new client: ", ls.b2s(buf))
			pc := NewProxyClient(conn, ls.b2s(buf))
			ls.addProxy(pc)
		}
	}
}

func (ls *Lrps) b2i(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}

func (ls *Lrps) b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
