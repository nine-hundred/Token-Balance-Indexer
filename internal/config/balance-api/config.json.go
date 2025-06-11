package balance_api

import (
	"encoding/json"
	"log"
	"onbloc/internal/config"
	"os"
)

type BalanceAPIConfig struct {
	Port int
	DB   config.Database
}

func Load(path string) (config BalanceAPIConfig, err error) {
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
