package main

import (
	"lrp/internal/client"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	c := client.NewClient("127.0.0.1:443", "192.168.3.239:8000")
	c.Run()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
}
