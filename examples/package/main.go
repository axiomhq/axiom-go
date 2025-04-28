package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/axiomhq/axiom-go/examples/package/log"
	"github.com/axiomhq/axiom-go/examples/package/service_a"
)

var logger = log.GetLogger("my-app")

func main() {
	logger.Info("started")

	service_a.Run()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("finished")
	log.Flush()
	os.Exit(0)
}
