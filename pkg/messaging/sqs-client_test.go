package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"onbloc/pkg/model"
	"testing"
)

func TestSQSClient_PublishEvent(t *testing.T) {
	sqsClient, err := NewSQSClient(context.TODO(), "http://localhost:4566/000000000000/test-queue")
	assert.Nil(t, err)
	err = sqsClient.put(context.TODO(), struct {
		Name string
	}{Name: "kcm"})
	assert.Nil(t, err)

}

func TestSQSClient_ReceiveMessage(t *testing.T) {
	sqsClient, err := NewSQSClient(context.TODO(), "http://localhost:4566/000000000000/test-queue")
	assert.Nil(t, err)
	msg, err := sqsClient.ReceiveMessage(context.TODO())
	assert.Nil(t, err)
	fmt.Println("msg:", msg.JsonData)
	var tokenEvent model.TokenEvent
	err = json.Unmarshal([]byte(msg.JsonData), &tokenEvent)
	assert.Nil(t, err)
	fmt.Println("after", tokenEvent)
}
