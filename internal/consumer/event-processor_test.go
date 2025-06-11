package consumer

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"onbloc/internal/repository/postgresdb"
	block_synchronizer "onbloc/internal/service/block-synchronizer"
	"onbloc/pkg/caching"
	"onbloc/pkg/messaging"
	"onbloc/pkg/model"
	"testing"
)

func TestEventProcessor_ProcessEvent(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=localhost user=postgres password=password dbname=onbloc port=5432 sslmode=disable"), &gorm.Config{})
	repository := postgresdb.NewRepository(db, 100)

	messageQueue, err := messaging.NewSQSClient(context.TODO(), "http://localhost:4566/000000000000/test-queue")
	assert.Nil(t, err)

	redis := caching.NewRedisClient("localhost:6379", "", 0)
	err = redis.Set(context.TODO(), "1", 1)
	assert.Nil(t, err)
	ep := NewEventProcessor(redis, messageQueue, repository)

	te := model.TokenEvent{
		Type:    block_synchronizer.EventTypeTransfer,
		PkgPath: "gno.land/r/gnoswap/v1/test_token/bar",
		Func:    block_synchronizer.EventFuncMint,
		To:      "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
		From:    "",
		Amount:  100000000000000,
	}
	err = ep.ProcessEvent(context.TODO(), te)
	assert.Nil(t, err)
}
