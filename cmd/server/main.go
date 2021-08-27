package main

import (
	"fmt"
	"lrp"
	"os"
	"os/signal"
	"syscall"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log := logrus.New()
	log.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})
	log.SetOutput(os.Stdout)
}

func main() {
	var pcs []*lrp.ProxyClient
	handleAdd := func(pc *lrp.ProxyClient) bool {
		fmt.Println("proxy client connected: ", pc.SiteId)
		pcs = append(pcs, pc)
		go func() {
			testOpenProxy(pc)
			time.Sleep(time.Second * 60)
			testCloseProxy(pc)
		}()
		return true
	}

	ls := lrp.NewLrps(log, ":8805", ":8806", handleAdd)
	if ok := ls.Run(); !ok {
		return
	}
	log.Info("ProxyServer start successfull")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
}

func testCloseProxy(pc *lrp.ProxyClient) {
	log.Info("test close proxy siteId: ", pc.SiteId)
	if err := pc.Close(); err != nil {
		log.Warn("close err:", err)
	} else {
		log.Info("Client Proxy Status:", pc.IsOpen)
	}
}

func testOpenProxy(pc *lrp.ProxyClient) {
	log.Info("start open client proxy: ", pc.SiteId)
	if err := pc.Open(); err != nil {
		log.Warn("open error", err)
	} else {
		log.Info("open client proxy succesfull")
		log.Info("site: ", pc.SiteId)
		log.Info("port: ", pc.Port)
		log.Info("status: ", pc.IsOpen)
	}
}
