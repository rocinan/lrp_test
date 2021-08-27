package lrp

import (
	"bytes"
	"encoding/binary"
	"lrp/internal/client"
	"net"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
)

type Lrpc struct {
	log        *logrus.Logger
	client     *client.Client
	server     string
	localAddr  string
	proxyPort  string
	notifyPort string
}

func (lc *Lrpc) Run(siteId string) bool {
	if siteId == "" {
		lc.log.Warn("siteId is required")
		return false
	}
	if conn, err := net.Dial("tcp4", lc.server+lc.notifyPort); err != nil {
		lc.log.Warn("connect server error: ", err)
		return false
	} else {
		id := lc.s2b(siteId)
		data := append(lc.i2b(len(id)), id...)
		if _, err := conn.Write(data); err != nil {
			lc.log.Warn("connect server error: ", err)
			conn.Close()
			return false
		}
		lc.log.Info("connect notify server successful")
		go func() {
			defer conn.Close()
			cmd := make([]byte, 1)
			for {
				if _, err := conn.Read(cmd); err != nil {
					lc.log.Warn("read server error: ", err)
					return
				} else {
					switch cmd[0] {
					case 0:
						if err := lc.client.Close(); err != nil {
							lc.log.Warn("close proxy error", err)
							conn.Write([]byte{0})
						} else {
							lc.log.Info("close proxy ok")
							conn.Write([]byte{1})
						}
					case 1:
						lc.log.Info("request open proxy")
						if client, ok := lc.startProxy(); !ok {
							conn.Write([]byte{0})
						} else {
							lc.log.Info("open proxy ok")
							lc.client = client
							data := append([]byte{1}, lc.i2b(client.ServerPort)...)
							if _, err := conn.Write(data); err != nil {
								lc.log.Warn("write to server error", err)
								return
							}
							lc.log.Info("request open proxy ok port: ", client.ServerPort)
						}
					default:
						lc.log.Warn("unknow cmd type")
						return
					}
				}
			}
		}()
	}
	return true
}

func (lc *Lrpc) startProxy() (*client.Client, bool) {
	c := client.NewClient(lc.localAddr, lc.server+lc.proxyPort)
	ch := make(chan bool)
	if err := c.Run(ch); err != nil {
		lc.log.Warn("connect proxy server fail: ", err)
		return nil, false
	} else {
		for {
			select {
			case init := <-ch:
				if init {
					return c, true
				} else {
					return nil, false
				}
			case <-time.After(time.Second * 15):
				return nil, false
			}
		}
	}
}

func (lc *Lrpc) s2b(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func (lc *Lrpc) i2b(n int) []byte {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}
