package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"paint-backend/internal/logger"
)

func main() {
	logger.Initialize()
	setupConfiguration()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}

func setupConfiguration() {
	viper.SetConfigName("configuration")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal(err.Error())
	}
}
