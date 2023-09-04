package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"paint-backend/internal/logger"
	"paint-backend/internal/repo"
)

var (
	dbConn        *pgx.Conn
	productsTable *repo.ProductsTable
)

func main() {
	logger.Initialize()
	setupConfiguration()
	setupDatabase()
	setupTables()

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
