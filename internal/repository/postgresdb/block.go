package postgresdb

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"onbloc/pkg/model"
	"time"
)

type Repository struct {
	db        *gorm.DB
	batchSize int
}

func NewRepository(db *gorm.DB, batchSize int) *Repository {
	return &Repository{db: db, batchSize: batchSize}
}

func (r Repository) GetLatestHeight(ctx context.Context) (int64, error) {
	var block model.Block
	err := r.db.WithContext(ctx).Order("height desc").First(&block).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return block.Height, nil
}

func (r Repository) InsertBlockRange(ctx context.Context, blocks []*model.Block) error {
	if len(blocks) == 0 {
		return nil
	}
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "height"}},
			DoNothing: true,
		}).CreateInBatches(blocks, len(blocks)).
		Error

	if err != nil {
		return err
	}
	return nil
}

func (r Repository) InsertTransactions(ctx context.Context, transactions []*model.BlockTransaction) error {
	if len(transactions) == 0 {
		return nil
	}
	err := r.db.WithContext(ctx).
		CreateInBatches(transactions, len(transactions)).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (r Repository) InsertTokenEvent(ctx context.Context, event model.TokenEvent) error {
	return r.db.WithContext(ctx).Create(&event).Error
}

func (r Repository) UpsertBalance(ctx context.Context, pkgPath, addr string, amount int64) error {
	balance := model.Balance{
		Address:   addr,
		TokenPath: pkgPath,
		Amount:    amount,
		UpdatedAt: time.Now(),
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "address"},
			{Name: "token_path"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"amount":     gorm.Expr("balances.amount + EXCLUDED.amount"),
			"updated_at": gorm.Expr("EXCLUDED.updated_at"),
		}),
	}).Create(&balance).Error
}

func (r Repository) UpdateTransferBalances(ctx context.Context, event model.TokenEvent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := Repository{
			db:        tx,
			batchSize: r.batchSize,
		}

		if err := txRepo.UpsertBalance(ctx, event.PkgPath, event.From, -event.Amount); err != nil {
			return err
		}

		return txRepo.UpsertBalance(ctx, event.PkgPath, event.To, event.Amount)
	})
}
