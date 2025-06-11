package tx_indexer

import (
	"context"
	"github.com/shurcooL/graphql"
	"net/http"
	"time"
)

type Client struct {
	client *graphql.Client
}

func NewClient(url string, timeout time.Duration) *Client {
	httpClient := &http.Client{Timeout: timeout}

	client := graphql.NewClient(url, httpClient)

	return &Client{client: client}
}

func (c *Client) GetBlocks(ctx context.Context, fromHeight, toHeight int64) (*GetBlocksResponse, error) {
	variables := map[string]interface{}{
		"gt": graphql.Int(fromHeight),
		"lt": graphql.Int(toHeight + 1),
	}

	var query struct {
		GetBlocks []Block `graphql:"getBlocks(where: {height: {gt: $gt, lt: $lt}})"`
	}
	err := c.client.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	return &GetBlocksResponse{
		Blocks: query.GetBlocks,
	}, nil
}

func (c Client) GetLatestBlockHeight(ctx context.Context) (int64, error) {
	var query struct {
		LatestBlockHeight int64 `graphql:"latestBlockHeight"`
	}

	err := c.client.Query(ctx, &query, nil)
	if err != nil {
		return 0, err
	}

	return query.LatestBlockHeight, nil
}

func (c *Client) GetTransactions(ctx context.Context, fromHeight, toHeight int64) (*GetTransactionsResponse, error) {
	var query struct {
		GetTransactions []Transaction `graphql:"getTransactions(where: {block_height: {gt: $gt, lt: $lt}, index: {lt: 1000}})"`
	}

	variables := map[string]interface{}{
		"gt": graphql.Int(fromHeight),
		"lt": graphql.Int(toHeight + 1),
	}

	err := c.client.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	return &GetTransactionsResponse{
		GetTransactions: query.GetTransactions,
	}, nil
}
