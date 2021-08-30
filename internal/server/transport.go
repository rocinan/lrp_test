package server

import (
	"context"
	"net"

	"github.com/lev2048/lrp/internal/utils"

	"github.com/lev2048/lrp/internal/conn"

	_cache "github.com/patrickmn/go-cache"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type Transport struct {
	nc       conn.Conn
	ln       net.Listener
	log      *logrus.Logger
	ctx      context.Context
	exit     context.CancelFunc
	connList *_cache.Cache
}

type Conn struct {
	nc   net.Conn
	exit chan bool
}

func NewTransport(nc conn.Conn, ln net.Listener, log *logrus.Logger) *Transport {
	ctx, cancel := context.WithCancel(context.Background())
	return &Transport{nc, ln, log, ctx, cancel, _cache.New(_cache.NoExpiration, _cache.NoExpiration)}
}

func (t *Transport) Process() {
	for {
		select {
		case <-t.ctx.Done():
			t.log.Info("transport exit")
			return
		default:
			if conn, err := t.ln.Accept(); err != nil {
				t.log.WithField("info", err).Warn("proxy server connect err")
				continue
			} else {
				t.log.Info("new client conn forward")
				go t.handleTransport(conn)
			}
		}
	}
}

func (t *Transport) handleTransport(conn net.Conn) {
	defer t.log.Info("conn transport exit")
	buf := make([]byte, 1460*3)
	cid := xid.New().Bytes()
	nc := &Conn{conn, make(chan bool)}
	t.connList.Set(utils.Bytes2Str(cid), nc, _cache.NoExpiration)
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-nc.exit:
			return
		default:
			if ln, err := conn.Read(buf); err != nil {
				utils.EncodeSend(t.nc, append([]byte{3, 0}, cid...))
				return
			} else {
				payload := append([]byte{2}, cid...)
				payload = append(payload, buf[:ln]...)
				if err := utils.EncodeSend(t.nc, payload); err != nil {
					t.log.WithField("info", err).Warn("write to client err close forward ...")
					return
				}
			}
		}
	}
}

func (t *Transport) Write(cid string, data []byte) bool {
	if conn, ok := t.connList.Get(cid); !ok {
		return false
	} else {
		if _, err := conn.(*Conn).nc.Write(data); err != nil {
			return false
		}
	}
	return true
}

func (t *Transport) CloseConn(cid string) bool {
	t.log.Info("forward close")
	if conn, ok := t.connList.Get(cid); !ok {
		return false
	} else {
		close(conn.(*Conn).exit)
		t.connList.Delete(cid)
	}
	return true
}
