package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"lrp/internal/server"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})
	log.SetOutput(os.Stdout)

	pcs := make([]*ProxyClient, 0)
	add := func(pc *ProxyClient) bool {
		pcs = append(pcs, pc)
		return true
	}

	ls := NewLrps(log, ":8805", ":8806", add)
	if ok := ls.Run(); !ok {
		return
	}

	log.Info("server start ok")
	time.Sleep(time.Second * 10)

	log.Info("start open client proxy: ", pcs[0].SiteId)
	if err := pcs[0].Open(); err != nil {
		log.Warn("open error", err)
	} else {
		log.Info("open proxy succesfull")
		log.Info("site: ", pcs[0].SiteId)
		log.Info("port: ", pcs[0].Port)
		log.Info("status: ", pcs[0].IsOpen)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig

	log.Info("close proxy")
	if err := pcs[0].Close(); err != nil {
		log.Warn("close err:", err)
	} else {
		log.Info("close ok", pcs[0].IsOpen)
	}

}

type Lrps struct {
	log        *logrus.Logger
	proxyPort  string
	notifyPort string
	addProxy   func(*ProxyClient) bool
}

func NewLrps(log *logrus.Logger, pp, np string, addFunc func(*ProxyClient) bool) *Lrps {
	return &Lrps{log, pp, np, addFunc}
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

type ProxyClient struct {
	conn   net.Conn
	lock   bool
	IsOpen bool
	Port   string
	SiteId string
}

func NewProxyClient(conn net.Conn, siteId string) *ProxyClient {
	return &ProxyClient{conn, false, false, "", siteId}
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
