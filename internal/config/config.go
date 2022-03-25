package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Secret string `json:"secret"`
}

func GetConfig() Config {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	if err := decoder.Decode(&config); err != nil {
		panic(err)
	}

	return config
}
