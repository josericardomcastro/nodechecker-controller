package main

import (
	"github.com/josericardomcastro/nodechecker-controller/config"
	"github.com/josericardomcastro/nodechecker-controller/controllers"
	"github.com/sirupsen/logrus"
)

var (
	log logrus.Entry
)

func init() {
	config.SetEnvConfig()
	config.SetLogConfig()
}

func main() {
	metrics := config.StartMetricServer()
	controllers.StartNodeCheckerController(metrics)
}
