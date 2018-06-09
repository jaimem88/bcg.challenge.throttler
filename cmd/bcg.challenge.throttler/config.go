package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// HTTP config
type HTTP struct {
	ListenPort string `json:"listen_port,omitempty"`
}

// Throttling config
type Throttling struct {
	N int64 `json:"n,omitempty"` // N number of requests per M milliseconds
	M int64 `json:"m,omitempty"` // M in milliseconds
}

var config = struct {
	Environment string      `json:"environment,omitempty"`
	HTTP        *HTTP       `json:"http,omitempty"`
	Throttling  *Throttling `json:"throttling,omitempty"`
}{
	Environment: "local",
	HTTP: &HTTP{
		ListenPort: os.Getenv("PORT"),
	},
	Throttling: &Throttling{
		N: 10,
		M: 1000,
	},
}

func writeDefaultConfig(location string) {
	f, err := os.Create(location)
	if err != nil {
		log.Fatalln("Couldn't open", location)
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	_, err = f.Write(data)
	if err != nil {
		log.Fatalln("Couldn't write to", location)
	}
}

func loadConfig(location string) {
	raw, err := ioutil.ReadFile(location)
	if err != nil {
		log.Fatalln("Couldn't open ", location)
	}

	err = json.Unmarshal(raw, &config)
	if err != nil {
		log.Fatalln("Couldn't understand config in", location, "-", err)
	}
}
