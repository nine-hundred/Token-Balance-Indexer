package block_synchronizer

import (
	"encoding/json"
	"log"
	"os"
)

type BlockSynchronizerConfig struct {
	TxIndexerEndPoint string `json:"txIndexerEndPoint"`
	BackFillBatchSize int    `json:"backFillBatchSize"`
	SyncInterval      int    `json:"syncInterval"`
	MessageQueueUrl   string `json:"messageQueueUrl"`
}

func Load(path string) (config BlockSynchronizerConfig, err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		log.Println(err)
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return
	}

	return
}
