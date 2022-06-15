package config

import (
	"github.com/sirupsen/logrus"
	"os"
)

func SetLogConfig() {
	// Log as JSON instead of the default ASCII formatter.
	if os.Getenv("ENVIRONMENT") == "production" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	// Set log level
	if os.Getenv("LOG_LEVEL") != "All" {
		switch level := os.Getenv("LOG_LEVEL"); level {
		case "Trace":
			logrus.SetLevel(logrus.TraceLevel)
		case "Debug":
			logrus.SetLevel(logrus.DebugLevel)
		case "Info":
			logrus.SetLevel(logrus.InfoLevel)
		case "Warning":
			logrus.SetLevel(logrus.WarnLevel)
		case "Error":
			logrus.SetLevel(logrus.ErrorLevel)
		case "Fatal":
			logrus.SetLevel(logrus.FatalLevel)
		case "Panic":
			logrus.SetLevel(logrus.PanicLevel)
		}
	}

	// Output to stdout instead of the default stderr
	logrus.SetOutput(os.Stdout)
}
