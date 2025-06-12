package block_synchronizer

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"onbloc/internal/repository/postgresdb"
	tx_indexer "onbloc/internal/tx-indexer"
	"onbloc/pkg/messaging"
	"os"
	"testing"
	"time"
)

var QueueUrl = "http://localhost:4566/000000000000/test-queue"

func TestService_SyncTransactionRage(t *testing.T) {
	client := tx_indexer.NewClient("https://dev-indexer.api.gnoswap.io/graphql/query", 30*time.Second)
	db, err := gorm.Open(postgres.Open("host=localhost user=postgres password=password dbname=onbloc port=5432 sslmode=disable"), &gorm.Config{})
	assert.Nil(t, err)

	repository := postgresdb.NewRepository(db)
	messageQueue, err := messaging.NewSQSClient(context.TODO(), QueueUrl)
	service := NewService(client, repository, messageQueue, 100, 5)

	err = service.SyncTransactionRage(context.TODO(), 667, 669)
	assert.Nil(t, err)

	defer messageQueue.CleanQueue(context.TODO(), QueueUrl)
}

func TestService_publishTransactionEvents(t *testing.T) {
	client := tx_indexer.NewClient("https://dev-indexer.api.gnoswap.io/graphql/query", 30*time.Second)
	db, err := gorm.Open(postgres.Open("host=localhost user=postgres password=password dbname=onbloc port=5432 sslmode=disable"), &gorm.Config{})
	assert.Nil(t, err)
	repository := postgresdb.NewRepository(db)
	messageQueue, err := messaging.NewSQSClient(context.TODO(), QueueUrl)
	assert.Nil(t, err)
	service := NewService(client, repository, messageQueue, 100, 5)

	var dummyTransactions tx_indexer.Transaction
	dummyData, err := os.ReadFile("./testDummy.json")
	assert.Nil(t, err)

	err = json.Unmarshal(dummyData, &dummyTransactions)
	assert.Nil(t, err)

	service.publishTransactionEvents(context.TODO(), []tx_indexer.Transaction{dummyTransactions})

	messageCount := service.messageQueue.GetMessageCount(context.TODO())
	assert.Equal(t, messageCount, 6)

	defer messageQueue.CleanQueue(context.TODO(), QueueUrl)
}
