package postgresdb

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"onbloc/pkg/model"
	"time"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
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

func (r Repository) WithTransaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r Repository) InsertTokenEventTx(ctx context.Context, tx *gorm.DB, event model.TokenEvent) (err error) {
	return tx.WithContext(ctx).Create(&event).Error
}

func (r Repository) InsertTokenEvent(ctx context.Context, event model.TokenEvent) error {
	return r.db.WithContext(ctx).Create(&event).Error
}

func (r Repository) UpsertBalance(ctx context.Context, tx *gorm.DB, pkgPath, addr string, amount int64) error {
	balance := model.Balance{
		Address:   addr,
		TokenPath: pkgPath,
		Amount:    amount,
		UpdatedAt: time.Now(),
	}

	return tx.WithContext(ctx).Clauses(clause.OnConflict{
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
			db: tx,
		}

		if err := txRepo.UpsertBalance(ctx, nil, event.PkgPath, event.From, -event.Amount); err != nil {
			return err
		}

		return txRepo.UpsertBalance(ctx, nil, event.PkgPath, event.To, event.Amount)
	})
}

func (r Repository) GetBalancesByAddress(ctx context.Context, addr string) (balances []model.Balance, err error) {
	err = r.db.WithContext(ctx).
		Where("address = ?", addr).Find(&balances).Error
	if err != nil {
		return nil, err
	}
	return
}

func (r Repository) GetAllBalances(ctx context.Context, offset, limit int) (balances []model.Balance, err error) {
	err = r.db.WithContext(ctx).
		Offset(offset).Limit(limit).Find(&balances).Error
	if err != nil {
		return nil, err
	}
	return
}

func (r Repository) GetTokenPathBalanceByAddress(ctx context.Context, tokenPath, address string) (balances []model.Balance, err error) {
	err = r.db.WithContext(ctx).
		Where("token_path = ? and address = ?", tokenPath, address).
		Find(&balances).Error
	if err != nil {
		return nil, err
	}
	return
}

func (r Repository) GetAllTokenPathBalances(ctx context.Context, tokenPath string) (balances []model.Balance, err error) {
	err = r.db.WithContext(ctx).
		Where("token_path = ?", tokenPath).
		Find(&balances).Error
	if err != nil {
		return nil, err
	}
	return
}

func (r Repository) GetTokenTransferHistoryByAddress(ctx context.Context, address string) (tokenEvents []model.TokenEvent, err error) {
	err = r.db.WithContext(ctx).
		Where("from_addr = ? or to_addr = ?", address, address).
		Find(&tokenEvents).Error
	if err != nil {
		return nil, err
	}
	return
}

func (r Repository) GetTokenTransferHistories(ctx context.Context) (tokenEvents []model.TokenEvent, err error) {
	err = r.db.WithContext(ctx).
		Find(&tokenEvents).Error
	if err != nil {
		return nil, err
	}
	return
}
