package tx_indexer

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	client := NewClient("https://dev-indexer.api.gnoswap.io/graphql/query", 30*time.Second)

	ctx := context.TODO()
	resp, err := client.GetTransactions(ctx, 0, 1000)
	log.Println(err)

	assert.NotEqual(t, 0, len(resp.GetTransactions))
	log.Println("len:", len(resp.GetTransactions))
	log.Println(resp.GetTransactions[0])
}
