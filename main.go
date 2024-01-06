package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
	"paint-backend/internal/endpoint"
	"paint-backend/internal/logger"
	"paint-backend/internal/repo"
	"paint-backend/internal/s3"
)

var (
	dbPool      *pgxpool.Pool
	httpHandler *endpoint.HttpHandler

	storage           *s3.Storage
	productsTable     *repo.ProductsTable
	subjectsTable     *repo.SubjectsTable
	brandsTable       *repo.BrandsTable
	subjectBrandTable *repo.SubjectBrandTable
)

func main() {
	logger.Initialize()
	setupConfiguration()
	setupDatabase()
	setupTables()
	setupStorage()

	httpHandler = endpoint.NewHttpHandler(storage, productsTable, subjectsTable, brandsTable, subjectBrandTable)
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

	dbPool.Close()
}

func setupConfiguration() {
	viper.SetConfigName("configuration")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal(err.Error())
	}
}

func setupDatabase() {
	dbUrl := viper.GetString("sql.url")

	var err error
	dbPool, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		logrus.Fatal("Failed to connect to db: ", err.Error())
	}

	err = dbPool.Ping(context.Background())
	if err != nil {
		logrus.Fatal("Failed to ping db: ", err.Error())
	}
}

func setupTables() {
	productsTable = repo.NewProductsTable(dbPool)
	subjectsTable = repo.NewSubjectsTable(dbPool)
	brandsTable = repo.NewBrandsTable(dbPool)
	subjectBrandTable = repo.NewSubjectBrandTable(dbPool)
}

func setupStorage() {
	storage = s3.NewStorage()
}
