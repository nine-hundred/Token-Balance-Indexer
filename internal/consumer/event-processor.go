package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"onbloc/internal/repository/postgresdb"
	block_synchronizer "onbloc/internal/service/block-synchronizer"
	"onbloc/pkg/caching"
	"onbloc/pkg/messaging"
	"onbloc/pkg/model"
)

type EventProcessor struct {
	caching         caching.Caching
	messageQueue    *messaging.SQSClient
	repository      *postgresdb.Repository
	eventStrategies map[string]EventStrategy
	batchSize       int
}

func NewEventProcessor(cache caching.Caching, messageQueue *messaging.SQSClient, repository *postgresdb.Repository, batchSize int) *EventProcessor {
	p := &EventProcessor{
		caching:      cache,
		messageQueue: messageQueue,
		repository:   repository,
		batchSize:    batchSize,
	}
	p.eventStrategies = map[string]EventStrategy{
		block_synchronizer.EventFuncTransfer: p.processTransferEvent,
		block_synchronizer.EventFuncBurn:     p.processBurnEvent,
		block_synchronizer.EventFuncMint:     p.processMintEvent,
	}
	return p
}

type EventStrategy func(ctx context.Context, event model.TokenEvent) error

func (p EventProcessor) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := p.consume(ctx); err != nil {
				log.Printf("consume err: %v", err)
			}
		}
	}
}

func (p EventProcessor) consume(ctx context.Context) error {
	message, err := p.messageQueue.ReceiveMessage(ctx)
	if err != nil {
		return err
	}
	if message.IsEmpty() {
		return nil
	}

	var tokenEvent model.TokenEvent
	err = json.Unmarshal([]byte(message.JsonData), &tokenEvent)
	if err != nil {
		p.messageQueue.DeleteMessage(ctx, message)
		return err
	}

	if err = p.ProcessEvent(ctx, tokenEvent); err != nil {
		log.Printf("Failed to process event: %v", err)
	}
	return p.messageQueue.DeleteMessage(ctx, message)
}

func (p EventProcessor) ProcessEvent(ctx context.Context, event model.TokenEvent) error {
	err := p.repository.InsertTokenEvent(ctx, event)
	if err != nil {
		log.Printf("Failed to insert token event: %v", err)
		return err
	}

	strategy, exists := p.eventStrategies[event.Func]
	if !exists {
		return fmt.Errorf("unsupported event function: %s", event.Func)
	}

	return strategy(ctx, event)
}

func (p EventProcessor) processMintEvent(ctx context.Context, event model.TokenEvent) error {
	return p.repository.UpsertBalance(ctx, event.PkgPath, event.To, event.Amount)
}

func (p EventProcessor) processBurnEvent(ctx context.Context, event model.TokenEvent) error {
	return p.repository.UpsertBalance(ctx, event.PkgPath, event.From, -event.Amount)
}

func (p EventProcessor) processTransferEvent(ctx context.Context, event model.TokenEvent) error {
	return p.repository.UpdateTransferBalances(ctx, event)
}
