package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"paint-backend/internal/logger"
	s3storage2 "paint-backend/internal/s3storage"
	"paint-backend/pkg/fserver"
)

func main() {
	logger.Initialize()
	setupConfiguration()

	fs := fserver.NewCommonFileServer(s3storage2.NewS3Storage(s3storage2.Config{
		BucketName: viper.GetString("s3.bucket"),
		Region:     viper.GetString("s3.region"),
		Host:       viper.GetString("s3.host"),
		Access:     os.Getenv(viper.GetString("s3.access")),
		Secret:     os.Getenv(viper.GetString("s3.secret")),
	}))

	list, err := fs.GetFilesList()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	logrus.Info(list)

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
