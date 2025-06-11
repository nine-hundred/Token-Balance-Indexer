package event_processor

import (
	"encoding/json"
	"log"
	"onbloc/internal/config"
	"os"
)

type EventProcessorConfig struct {
	MessageQueueUrl string       `json:"messageQueueUrl"`
	Caching         config.Redis `json:"caching"`
}

func Load(path string) (config EventProcessorConfig, err error) {
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
