package consumer

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"onbloc/internal/repository/postgresdb"
	block_synchronizer "onbloc/internal/service/block-synchronizer"
	"onbloc/pkg/caching"
	"onbloc/pkg/messaging"
	"onbloc/pkg/model"
	"testing"
)

func TestEventProcessor_ProcessEvent(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=localhost user=postgres password=password dbname=onbloc port=5432 sslmode=disable"), &gorm.Config{
		Logger: logger.Default,
	})
	repository := postgresdb.NewRepository(db)

	messageQueue, err := messaging.NewSQSClient(context.TODO(), "http://localhost:4566/000000000000/test-queue")
	assert.Nil(t, err)

	redis := caching.NewRedisClient("localhost:6379", "", 0)

	ep := NewEventProcessor(redis, messageQueue, repository, 0)

	transactionHash := "Madp4C64dGZV4zrrNrz1HduBNa7yDBZRr544oNv39e4"
	t.Run("mint 이벤트 처리", func(t *testing.T) {
		te := model.TokenEvent{
			Type:            block_synchronizer.EventTypeTransfer,
			TransactionHash: transactionHash,
			TxEventIndex:    1,
			PkgPath:         "gno.land/r/gnoswap/v1/test_token/bar",
			Func:            block_synchronizer.EventFuncMint,
			To:              "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
			From:            "",
			Amount:          100000000000000,
		}
		err = ep.ProcessEvent(context.TODO(), te)
	})

	t.Run("burn 이벤트 처리", func(t *testing.T) {
		te := model.TokenEvent{
			Type:            block_synchronizer.EventTypeTransfer,
			TransactionHash: transactionHash,
			TxEventIndex:    2,
			PkgPath:         "gno.land/r/gnoswap/v1/test_token/bar",
			Func:            block_synchronizer.EventFuncBurn,
			To:              "",
			From:            "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
			Amount:          100000000000000,
		}
		err = ep.ProcessEvent(context.TODO(), te)
		assert.Nil(t, err)
	})

	t.Run("Transfer 이벤트 처리", func(t *testing.T) {
		te := model.TokenEvent{
			Type:            block_synchronizer.EventTypeTransfer,
			TransactionHash: transactionHash,
			TxEventIndex:    3,
			PkgPath:         "gno.land/r/gnoswap/v1/test_token/bar",
			Func:            block_synchronizer.EventFuncMint,
			To:              "g1cceshmzzlmrh7rr3z30j2t5mrvsq9yccysw9nu",
			From:            "",
			Amount:          100000000000000,
		}
		err = ep.ProcessEvent(context.TODO(), te)
		assert.Nil(t, err)
	})

	defer func() {
		db.Exec("delete from token_events where transaction_hash = ?", transactionHash)
	}()
}
