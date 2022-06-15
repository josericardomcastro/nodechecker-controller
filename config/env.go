package config

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func SetEnvConfig() {
	e := godotenv.Load()
	if e != nil {
		logrus.Info("No '.env' file found in local.")
	} else {
		logrus.Info(".env file loaded")
	}
}
