package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"log"
	"strconv"
	"time"
)

type MessageObject struct {
	ReceiptHandle *string
	CreatedTime   time.Time `json:"createdTime"`
	JsonData      string    `json:"jsonData"`
}

func (o MessageObject) IsEmpty() bool {
	if o.ReceiptHandle == nil {
		return true
	}
	return false
}

type SQSClient struct {
	url    string
	client *sqs.Client
}

func NewSQSClient(ctx context.Context, url string) (*SQSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	cfg.BaseEndpoint = &url
	sqsClient := sqs.NewFromConfig(cfg)
	return &SQSClient{
		client: sqsClient,
	}, nil
}

func (p SQSClient) PublishMessage(ctx context.Context, event interface{}) error {
	return p.put(ctx, event)
}

func (p SQSClient) put(ctx context.Context, obj interface{}) error {
	mo := MessageObject{}
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	mo.JsonData = string(data)
	mo.CreatedTime = time.Now()

	data, err = json.Marshal(mo)
	if err != nil {
		return fmt.Errorf("failed to marshal message object: %w", err)
	}

	messageBody := string(data)

	message := sqs.SendMessageInput{
		MessageBody: &messageBody,
		QueueUrl:    &p.url,
	}

	_, err = p.client.SendMessage(ctx, &message)
	if err != nil {
		log.Printf("Failed to send SQS message: %v", err)
		return fmt.Errorf("failed to send SQS message: %w", err)
	}
	return nil
}

func (p SQSClient) GetMessageCount(ctx context.Context) int {
	result, err := p.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: &p.url,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
		},
	})
	if err != nil {
		return 0
	}

	messageCountStr, exists := result.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)]
	if !exists {
		return 0
	}

	messageCount := 0
	messageCount, _ = strconv.Atoi(messageCountStr)
	return messageCount
}

func (p SQSClient) CleanQueue(ctx context.Context, url string) error {
	_, err := p.client.PurgeQueue(ctx, &sqs.PurgeQueueInput{QueueUrl: &url})
	return err
}

func (p SQSClient) ReceiveMessage(ctx context.Context) (message MessageObject, err error) {
	messages, err := p.get(ctx, 1)
	if err != nil {
		return MessageObject{}, err
	}

	if len(messages) > 0 {
		message = messages[0]
	}
	return
}

func (p SQSClient) get(ctx context.Context, count int32) (attrs []MessageObject, err error) {
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            &p.url,
		MaxNumberOfMessages: count,
	}

	resp, err := p.client.ReceiveMessage(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to receive SQS message: %w", err)
	}

	var messages []MessageObject
	for _, msg := range resp.Messages {
		var msgObj MessageObject
		err = json.Unmarshal([]byte(*msg.Body), &msgObj)
		if err != nil {
			log.Printf("fail to unmarshal message: %v", err)
		}
		msgObj.ReceiptHandle = msg.ReceiptHandle
		messages = append(messages, msgObj)
	}
	return messages, nil
}

func (p SQSClient) DeleteMessage(ctx context.Context, message MessageObject) error {
	_, err := p.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &p.url,
		ReceiptHandle: message.ReceiptHandle,
	})
	return err
}
