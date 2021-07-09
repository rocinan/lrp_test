package main

import (
	"context"
	"lrp/internal/server"
	"os"
	"os/signal"
	"syscall"

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
	addr := ":8000"
	ctx, cancel := context.WithCancel(context.Background())
	server := server.NewLrpServer(log, addr, ctx)
	if err := server.Run(); err != nil {
		log.WithField("info", err).Warn("lrps启动失败")
	}
	log.Info("server running ..." + addr)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	cancel()
}
