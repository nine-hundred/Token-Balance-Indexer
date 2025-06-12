package model

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type Block struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Hash      string    `json:"hash" gorm:"unique;not null;size:64"`
	Height    int64     `json:"height" gorm:"unique;not null;index"`
	Time      time.Time `json:"time" gorm:"not null;index"`
	NumTxs    int       `json:"num_txs" gorm:"not null;default:0"`
	TotalTxs  int64     `json:"total_txs" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type BlockTransaction struct {
	ID          int64           `gorm:"primaryKey"`
	IndexNum    int64           `gorm:"column:index_num"`
	Hash        string          `gorm:"column:hash;unique"`
	BlockHeight int64           `gorm:"column:block_height"`
	Success     bool            `gorm:"column:success"`
	GasWanted   int64           `gorm:"column:gas_wanted"`
	GasUsed     int64           `gorm:"column:gas_used"`
	Memo        string          `gorm:"column:memo"`
	GasFee      json.RawMessage `gorm:"column:gas_fee;type:jsonb"`
	Messages    json.RawMessage `gorm:"column:messages;type:jsonb"`
	Response    json.RawMessage `gorm:"column:response;type:jsonb"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
}

func (BlockTransaction) TableName() string {
	return "transactions"
}

type TokenEvent struct {
	TransactionHash string `json:"transactionHash" gorm:"column:transaction_hash;not null"`
	TxEventIndex    int    `json:"TxEventIndex" gorm:"column:tx_event_index; not null"`
	Type            string `json:"type" gorm:"column:type;not null"`
	PkgPath         string `json:"pkg_path" gorm:"column:pkg_path;not null"`
	Func            string `json:"func" gorm:"column:func;not null"`
	From            string `json:"from" gorm:"column:from_addr;not null"`
	To              string `json:"to" gorm:"column:to_addr; not null"`
	Amount          int64  `json:"amount" gorm:"column:amount; not null"`
}

type TokenEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TokenEventAttributes []TokenEventAttribute

func (t TokenEventAttributes) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return json.Marshal(t)
}

type Balance struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Address   string    `gorm:"type:varchar(255);not null;uniqueIndex:uk_balances_address_token" json:"address"`
	TokenPath string    `gorm:"type:varchar(255);not null;uniqueIndex:uk_balances_address_token;column:token_path" json:"token_path"`
	Amount    int64     `gorm:"type:bigint;not null;default:0" json:"amount"`
	CreatedAt time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:now()" json:"updated_at"`
}

func (b Balance) TableName() string {
	return "balances"
}
