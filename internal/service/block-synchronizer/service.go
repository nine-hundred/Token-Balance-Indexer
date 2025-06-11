package block_synchronizer

import (
	"context"
	"fmt"
	"log"
	"onbloc/internal/repository/postgresdb"
	"onbloc/internal/tx-indexer"
	"onbloc/pkg/messaging"
	"time"
)

type Service struct {
	backFillBatchSize int
	syncInterval      time.Duration
	indexerClient     *tx_indexer.Client
	repository        *postgresdb.Repository
	messageQueue      *messaging.SQSClient
}

func NewService(client *tx_indexer.Client, repository *postgresdb.Repository, queue *messaging.SQSClient, backFillBatchSize int, syncInterval time.Duration) *Service {
	return &Service{
		indexerClient:     client,
		repository:        repository,
		messageQueue:      queue,
		backFillBatchSize: backFillBatchSize,
		syncInterval:      syncInterval,
	}
}

func (s Service) RunBackFill(ctx context.Context) error {
	return s.runBackFill(ctx)
}

func (s Service) GetLatestHeight(ctx context.Context) (int64, error) {
	return s.indexerClient.GetLatestBlockHeight(ctx)
}

func (s Service) RunRealtimeSync(ctx context.Context) error {
	ticker := time.NewTicker(s.syncInterval * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("sync done")
			return ctx.Err()
		case <-ticker.C:
			lastProcessedHeight, currentBlockHeight, err := s.getHeightGap(ctx)
			if err != nil {
				log.Println(err)
				continue
			}

			if lastProcessedHeight < currentBlockHeight {
				for current := lastProcessedHeight; current < currentBlockHeight; {
					end := current + int64(s.backFillBatchSize)
					if end > currentBlockHeight {
						end = currentBlockHeight
					}
					err = s.SyncBlockRange(ctx, lastProcessedHeight, currentBlockHeight)
					if err != nil {
						log.Println("fail to sync block: ", err)
						continue
					}

					err = s.SyncTransactionRage(ctx, lastProcessedHeight, currentBlockHeight)
					if err != nil {
						log.Println("fail to sync block: ", err)
						continue
					}
					current = end + 1
				}
			}
		}
	}
}

func (s Service) getHeightGap(ctx context.Context) (lastProcessed, current int64, err error) {
	lastProcessed, err = s.repository.GetLatestHeight(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("fail to get height from db: %w", err)
	}

	current, err = s.GetLatestHeight(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("fail to get height from graphql: %w", err)
	}

	return lastProcessed, current, nil
}

func (s Service) runBackFill(ctx context.Context) error {
	for {
		lastProcessedHeight, currentBlockHeight, err := s.getHeightGap(ctx)
		if err != nil {
			return err
		}

		if lastProcessedHeight >= currentBlockHeight {
			log.Println("backFill 종료")
			break
		}

		start := lastProcessedHeight
		end := lastProcessedHeight + int64(s.backFillBatchSize)
		if end > currentBlockHeight {
			end = currentBlockHeight
		}

		err = s.SyncBlockRange(ctx, start, end)
	}
	return nil
}

func (s Service) SyncBlockRange(ctx context.Context, fromHeight, toHeight int64) error {
	// graphql에서 가져오기.
	resp, err := s.indexerClient.GetBlocks(ctx, fromHeight, toHeight)
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("save block start:%d, end:%d, block-len:%d", fromHeight, toHeight, len(resp.Blocks)))

	//db에 저장.
	err = s.repository.InsertBlockRange(ctx, resp.ToModels())
	if err != nil {
		return err
	}
	return nil
}

func (s Service) SyncTransactionRage(ctx context.Context, fromHeight, toHeight int64) error {
	resp, err := s.indexerClient.GetTransactions(ctx, fromHeight, toHeight)
	if err != nil {
		return fmt.Errorf("failed to get transactions from %d to %d: %w", fromHeight, toHeight, err)
	}

	transactions := resp.ToModels()
	err = s.repository.InsertTransactions(ctx, transactions)
	if err != nil {
		return fmt.Errorf("failed to insert transactions: %w", err)
	}
	log.Printf("save transactions start:%d, end:%d, transaction-len:%d\n", fromHeight, toHeight, len(resp.GetTransactions))

	s.publishTransactionEvents(ctx, resp.GetTransactions)

	return nil
}

func (s Service) publishTransactionEvents(ctx context.Context, transactions []tx_indexer.Transaction) {
	for _, transaction := range transactions {
		if len(transaction.Response.Events) == 0 {
			break
		}
		for _, event := range transaction.Response.Events {
			if !s.isTransferTokenEvent(event) {
				continue
			}
			err := s.messageQueue.PublishMessage(ctx, event.ToModel())
			if err != nil {
				log.Printf("fail to publish event: %v\n", event)
			}
		}
	}
}

const (
	EventTypeTransfer = "Transfer"
	EventFuncTransfer = "Transfer"
	EventFuncMint     = "Mint"
	EventFuncBurn     = "Burn"
)

type attrRequirement struct {
	fromEmpty bool
	toEmpty   bool
}

var attrRequirements = map[string]attrRequirement{
	EventFuncMint:     {fromEmpty: true, toEmpty: false},
	EventFuncBurn:     {fromEmpty: false, toEmpty: true},
	EventFuncTransfer: {fromEmpty: false, toEmpty: false},
}

func (s Service) isTransferTokenEvent(event tx_indexer.Event) bool {
	if event.GnoEvent.Type != EventTypeTransfer {
		return false
	}

	rule, exists := attrRequirements[event.GnoEvent.Func]
	if !exists {
		return false
	}
	return s.validateEventAttrs(event, rule.fromEmpty, rule.toEmpty)
}

func (s Service) validateEventAttrs(event tx_indexer.Event, fromEmpty, toEmpty bool) bool {
	attrs := event.GetAttrs()
	if len(attrs) != 3 {
		return false
	}

	return (attrs["from"] == "") == fromEmpty && (attrs["to"] == "") == toEmpty && attrs["value"] != ""
}
