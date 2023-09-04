package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
	"paint-backend/internal/endpoint"
	"paint-backend/internal/logger"
	"paint-backend/internal/repo"
	"paint-backend/internal/s3storage"
)

var (
	dbConn      *pgx.Conn
	httpHandler *endpoint.HttpHandler

	storage       *s3storage.Storage
	productsTable *repo.ProductsTable
)

func main() {
	logger.Initialize()
	setupConfiguration()
	setupDatabase()
	setupTables()
	setupStorage()

	httpHandler = endpoint.NewHttpHandler(storage, productsTable)
	go func() {
		logrus.Info("Server was started")
		err := fasthttp.ListenAndServe("0.0.0.0:8000", httpHandler.Handle)
		if err != nil {
			logrus.Warn(err.Error())
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	dbConn.Close(context.Background())
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

func setupDatabase() {
	dbUrl := viper.GetString("sql.url")

	var err error
	dbConn, err = pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		logrus.Fatal("Failed to connect to db: ", err.Error())
	}

	err = dbConn.Ping(context.Background())
	if err != nil {
		logrus.Fatal("Failed to ping db: ", err.Error())
	}
}

func setupTables() {
	var err error
	productsTable, err = repo.NewProductsTable(dbConn)
	if err != nil {
		logrus.Fatal("Failed to init products table: ", err.Error())
	}
}

func setupStorage() {
	storage = s3storage.NewStorage()
}
