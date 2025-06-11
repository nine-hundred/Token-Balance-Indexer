package tx_indexer

import (
	"encoding/json"
	"log"
	"onbloc/pkg/model"
	"strconv"
	"time"
)

type GetBlocksResponse struct {
	Blocks []Block
}

func (g GetBlocksResponse) ToModels() []*model.Block {
	blocks := make([]*model.Block, len(g.Blocks))
	for i, block := range g.Blocks {
		blocks[i] = block.ToModel()
	}
	return blocks
}

type Block struct {
	Hash     string    `graphql:"hash"`
	Height   int64     `graphql:"height"`
	Time     time.Time `graphql:"time"`
	NumTxs   int       `graphql:"num_txs"`
	TotalTxs int64     `graphql:"total_txs"`
}

func (b Block) ToModel() *model.Block {
	return &model.Block{
		Hash:     b.Hash,
		Height:   b.Height,
		Time:     b.Time,
		NumTxs:   b.NumTxs,
		TotalTxs: b.TotalTxs,
	}
}

type BlockWhereInput struct {
	Height *HeightCondition `json:"height,omitempty"`
}

type HeightCondition struct {
	GT *int `json:"gt,omitempty"`
	LT *int `json:"lt,omitempty"`
}

type GetTransactionsResponse struct {
	GetTransactions []Transaction
}

func (r GetTransactionsResponse) ToModels() []*model.BlockTransaction {
	transactions := make([]*model.BlockTransaction, len(r.GetTransactions))
	for i, transaction := range r.GetTransactions {
		trx, err := transaction.ToModel()
		if err != nil {
			log.Printf("fail to convert transaction at %s, %d\n", transaction.Hash, transaction.BlockHeight)
			continue
		}
		transactions[i] = trx
	}

	return transactions
}

type Transaction struct {
	Index       int64               `graphql:"index"`
	Hash        string              `graphql:"hash"`
	Success     bool                `graphql:"success"`
	BlockHeight int64               `graphql:"block_height"`
	GasWanted   int64               `graphql:"gas_wanted"`
	GasUsed     int64               `graphql:"gas_used"`
	Memo        string              `graphql:"memo"`
	GasFee      GasFee              `graphql:"gas_fee"`
	Messages    []Message           `graphql:"messages"`
	Response    TransactionResponse `graphql:"response"`
}

func (t Transaction) ToModel() (*model.BlockTransaction, error) {
	gasFeeBytes, err := json.Marshal(t.GasFee)
	if err != nil {
		return nil, err
	}

	messageBytes, err := json.Marshal(t.Messages)
	if err != nil {
		return nil, err
	}

	responseBytes, err := json.Marshal(t.Response)
	if err != nil {
		return nil, err
	}

	return &model.BlockTransaction{
		IndexNum:    t.Index,
		Hash:        t.Hash,
		BlockHeight: t.BlockHeight,
		Success:     t.Success,
		GasWanted:   t.GasWanted,
		GasUsed:     t.GasUsed,
		Memo:        t.Memo,
		GasFee:      gasFeeBytes,
		Messages:    messageBytes,
		Response:    responseBytes,
		CreatedAt:   time.Time{},
	}, nil
}

type GasFee struct {
	Amount int64  `graphql:"amount"`
	Denom  string `graphql:"denom"`
}

type Message struct {
	Route   string       `graphql:"route"`
	TypeUrl string       `graphql:"typeUrl"`
	Value   MessageValue `graphql:"value"`
}

type MessageValue struct {
	BankMsgSend   `graphql:"... on BankMsgSend"`
	MsgCall       `graphql:"... on MsgCall"`
	MsgAddPackage `graphql:"... on MsgAddPackage"`
	MsgRun        `graphql:"... on MsgRun"`
}

type BankMsgSend struct {
	FromAddress string `graphql:"from_address"`
	ToAddress   string `graphql:"to_address"`
	Amount      string `graphql:"amount"`
}

type MsgCall struct {
	Caller  string   `graphql:"caller"`
	Send    string   `graphql:"send"`
	PkgPath string   `graphql:"pkg_path"`
	Func    string   `graphql:"func"`
	Args    []string `graphql:"args"`
}

type MsgAddPackage struct {
	Creator string     `graphql:"creator"`
	Deposit string     `graphql:"deposit"`
	Package MemPackage `graphql:"package"`
}

type MsgRun struct {
	Caller  string     `graphql:"caller"`
	Send    string     `graphql:"send"`
	Package MemPackage `graphql:"package"`
}

type MemPackage struct {
	Name  string `graphql:"name"`
	Path  string `graphql:"path"`
	Files []File `graphql:"files"`
}

type File struct {
	Name string `graphql:"name"`
	Body string `graphql:"body"`
}

type TransactionResponse struct {
	Log    string  `graphql:"log"`
	Info   string  `graphql:"info"`
	Error  string  `graphql:"error"`
	Data   string  `graphql:"data"`
	Events []Event `graphql:"events"`
}

type Event struct {
	GnoEvent `graphql:"... on GnoEvent"`
}

type GnoEvent struct {
	Type    string      `json:"type" graphql:"type"`
	Func    string      `json:"func" graphql:"func"`
	PkgPath string      `json:"pkg_path" graphql:"pkg_path"`
	Attrs   []Attribute `json:"attrs" graphql:"attrs"`
}

func (e *Event) ToModel() *model.TokenEvent {
	tokenEvent := &model.TokenEvent{
		Type:    e.Type,
		PkgPath: e.PkgPath,
		Func:    e.Func,
	}
	for _, attr := range e.Attrs {
		switch attr.Key {
		case "from":
			tokenEvent.From = attr.Value
		case "to":
			tokenEvent.To = attr.Value
		case "value":
			if amount, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
				tokenEvent.Amount = amount
			}
		}
	}

	return tokenEvent
}

type Attribute struct {
	Key   string `graphql:"key"`
	Value string `graphql:"value"`
}
