package server

import (
	"context"
	"errors"
	"lrp/internal/conn"
	"lrp/internal/utils"
	"net"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type Server struct {
	ln        string
	log       *logrus.Logger
	ctx       context.Context
	netServer conn.Conn
}

func NewLrpServer(log *logrus.Logger, ln string, ctx context.Context) *Server {
	return &Server{ln, log, ctx, nil}
}

func (s *Server) Run() (err error) {
	ns, _ := conn.NewConn("tcp")
	if s.netServer, err = ns.Listen(s.ln); err != nil {
		return
	} else {
		go func() {
			for {
				select {
				case <-s.ctx.Done():
					err = errors.New("lrps get exit sig")
					return
				default:
					if netConn, err := s.netServer.Accept(); err != nil {
						s.log.WithField("info", err).Warn("client connect err")
						continue
					} else {
						go s.handleClientConn(netConn)
						s.log.Info("客户端连接成功")
					}
				}
			}
		}()
	}
	return
}

func (s *Server) handleClientConn(nc conn.Conn) {
	defer nc.Close()
	tr := new(Transport)
	for {
		payload, err := utils.DecodeReceive(nc)
		if err != nil {
			s.log.Info("client close")
			return
		}
		switch payload[0] {
		case 0:
			if ln, err := net.Listen("tcp", ":0"); err != nil {
				utils.EncodeSend(nc, []byte{1, 0})
				s.log.Warn("listen error server exit")
				return
			} else {
				tr = NewTransport(nc, ln, s.log)
				go tr.Process()
				bindPort, _ := strconv.Atoi(strings.Split(ln.Addr().String(), "[::]:")[1])
				if err := utils.EncodeSend(nc, append([]byte{1, 1}, utils.IntToBytes(bindPort)...)); err != nil {
					s.log.WithField("info", err).Warn("write client err")
					return
				}
			}
		case 2:
			if ok := tr.Write(utils.Bytes2Str(payload[1:13]), payload[13:]); !ok {
				s.log.Warn("forward data err , close transport")
				return
			}
		case 3:
			if payload[1] == 0 {
				if ok := tr.CloseConn(utils.Bytes2Str(payload[2:14])); !ok {
					s.log.Warn("delete conn error")
				}
			}
		default:
			s.log.Warn("cmd not supported")
			return
		}
	}
}
