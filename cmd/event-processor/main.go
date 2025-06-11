package main

import (
	"context"
	"flag"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	event_processor_config "onbloc/internal/config/event-processor"
	"onbloc/internal/consumer"
	"onbloc/internal/repository/postgresdb"
	"onbloc/pkg/caching"
	"onbloc/pkg/messaging"
)

func main() {
	path := ""
	flag.StringVar(&path, "c", "config.json", "config path")
	flag.Parse()

	conf, err := event_processor_config.Load(path)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	messageQueue, err := messaging.NewSQSClient(context.TODO(), conf.MessageQueueUrl)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to message queue: %s\n", err.Error()))
	}

	redis := caching.NewRedisClient(conf.Caching.GetAddr(), conf.Caching.Password, conf.Caching.DB)

	dsn := "host=localhost user=postgres password=password dbname=onbloc port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	repository := postgresdb.NewRepository(db)

	eventProcessor := consumer.NewEventProcessor(redis, messageQueue, repository, conf.BatchSize)
	err = eventProcessor.Start(context.Background())
	if err != nil {

	}
}
