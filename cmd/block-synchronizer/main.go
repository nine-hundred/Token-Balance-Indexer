package main

import (
	"context"
	"flag"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	block_synchronizer_config "onbloc/internal/config/block-synchronizer"
	"onbloc/internal/repository/postgresdb"
	block_synchronizer "onbloc/internal/service/block-synchronizer"
	"onbloc/internal/tx-indexer"
	"onbloc/pkg/messaging"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	path := ""
	flag.StringVar(&path, "c", "config.json", "config path")
	flag.Parse()

	conf, err := block_synchronizer_config.Load(path)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	client := tx_indexer.NewClient(conf.TxIndexerEndPoint, time.Second*60)
	dsn := "host=localhost user=postgres password=password dbname=onbloc port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %s\n", err.Error()))
	}
	repository := postgresdb.NewRepository(db, conf.BackFillBatchSize)
	messageQueue, err := messaging.NewSQSClient(context.TODO(), conf.MessageQueueUrl)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to message queue: %s\n", err.Error()))
	}
	service := block_synchronizer.NewService(client, repository, messageQueue, conf.BackFillBatchSize, time.Duration(conf.SyncInterval))

	err = service.RunBackFill(context.Background())
	if err != nil {
		panic(err)
	}
	log.Println("back-fill done")

	go service.RunRealtimeSync(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}
