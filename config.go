package main

import (
	"encoding/json"
	"os"
)

type config struct {
	Secret string `json:"secret"`
}

func getConfig() config {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := config{}
	if err := decoder.Decode(&config); err != nil {
		panic(err)
	}

	return config
}
