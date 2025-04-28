package service_a

import "github.com/axiomhq/axiom-go/examples/package/log"

var logger = log.GetLogger("service-a")

func Run() {
	logger.Info("running")
}
