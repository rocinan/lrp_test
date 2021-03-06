package client

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/lev2048/lrp/internal/conn"
	"github.com/lev2048/lrp/internal/utils"

	nested "github.com/antonfisher/nested-logrus-formatter"
	_cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type Client struct {
	laddr      string
	server     string
	ServerPort int
	conn       conn.Conn
	ctx        context.Context
	log        *logrus.Logger
	cancel     context.CancelFunc
	connlist   *_cache.Cache
}

func NewClient(laddr, server string) *Client {
	log := logrus.New()
	log.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})
	log.SetOutput(os.Stdout)
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{laddr, server, 0, nil, ctx, log, cancel, _cache.New(_cache.NoExpiration, _cache.NoExpiration)}
}

func (c *Client) Run(ch chan bool) error {
	if conn, err := conn.NewConn("tcp"); err != nil {
		return err
	} else {
		c.conn = conn
		if sc, err := conn.Dial(c.server); err != nil {
			return err
		} else {
			c.log.Info("server connected wait server reply ...")
			if err := utils.EncodeSend(sc, []byte{0}); err != nil {
				c.log.Warn("send request data error")
				return err
			}
			go c.handleServerData(sc, ch)
		}
	}
	return nil
}

func (c *Client) Close() error {
	c.cancel()
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Client) handleServerData(sc conn.Conn, ch chan bool) {
	defer c.cancel()
	for {
		payload, err := utils.DecodeReceive(sc)
		if err != nil {
			c.log.WithField("info", err).Warn("Link broken")
			return
		}
		switch payload[0] {
		case 1:
			if payload[1] == 0 {
				c.log.Warn("server bind port error")
				ch <- false
				return
			} else {
				c.ServerPort = utils.BytesToInt(payload[2:])
				ch <- true
				c.log.Info("connect successful")
				c.log.Info("clientAddr: " + c.laddr)
				c.log.Info("serverAddr: " + strings.Split(c.server, ":")[0] + ":" + strconv.Itoa(c.ServerPort))
			}
		case 2:
			cid := payload[1:13]
			tr, ok := c.connlist.Get(utils.Bytes2Str(cid))
			if !ok {
				tr = NewTransport(c.laddr, cid, sc, c.ctx, c.log)
				if err := tr.(*Transport).Process(); err != nil {
					c.log.WithField("err", err).Warn("create transport err")
					c.log.Warn(tr)
				}
				c.connlist.Set(utils.Bytes2Str(cid), tr, _cache.NoExpiration)
				c.log.Info("new transport created")
			}
			tr.(*Transport).Write(payload[13:])
		case 3:
			if payload[1] == 0 {
				tr, ok := c.connlist.Get(utils.Bytes2Str(payload[2:14]))
				if ok {
					tr.(*Transport).Close()
				}
			} else {
				c.log.Info("server close link")
				return
			}
		default:
			c.log.Warn("cmd not supported", payload)
			return
		}
	}
}
