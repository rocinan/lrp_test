package client

import (
	"context"
	"net"

	"github.com/lev2048/lrp/internal/conn"
	"github.com/lev2048/lrp/internal/utils"

	"github.com/sirupsen/logrus"
)

type Transport struct {
	sc    conn.Conn
	conn  net.Conn
	cid   []byte
	laddr string
	exit  chan bool
	log   *logrus.Logger
	ctx   context.Context
}

func NewTransport(lr string, cid []byte, sc conn.Conn, ctx context.Context, log *logrus.Logger) *Transport {
	return &Transport{
		sc:    sc,
		cid:   cid,
		log:   log,
		ctx:   ctx,
		laddr: lr,
		exit:  make(chan bool),
	}
}

func (t *Transport) Process() (err error) {
	if t.conn, err = net.Dial("tcp", t.laddr); err != nil {
		return
	} else {
		go func() {
			buf := make([]byte, 1460*3)
			for {
				select {
				case <-t.ctx.Done():
					return
				case <-t.exit:
					return
				default:
					if ln, err := t.conn.Read(buf); err != nil {
						utils.EncodeSend(t.sc, append([]byte{3, 0}, t.cid...))
						return
					} else {
						payload := append([]byte{2}, t.cid...)
						payload = append(payload, buf[:ln]...)
						if err := utils.EncodeSend(t.sc, payload); err != nil {
							t.log.WithField("info", err).Warn("write to server err")
							return
						}
					}
				}
			}
		}()
	}
	return
}

func (t *Transport) Write(data []byte) bool {
	if _, err := t.conn.Write(data); err != nil {
		t.log.Warn("write to local error")
		return false
	}
	return true
}

func (t *Transport) Close() {
	close(t.exit)
	t.conn.Close()
}
