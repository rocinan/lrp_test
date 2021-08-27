package main

import (
	"lrp"
	"os"
	"os/signal"
	"syscall"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})
	log.SetOutput(os.Stdout)
}

func main() {
	siteId := "LZX081f7108d927"

	lc := lrp.NewLrpc(log, "192.168.3.140", "127.0.0.1:80", ":8805", ":8806")
	if ok := lc.Run(siteId); !ok {
		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
}
