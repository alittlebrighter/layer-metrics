package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

var defaultConfig = &AppConfig{PollRate: 1, InfluxEndpoint: "localhost"}

// parseConfig reads from a configuration file (default ./config.json, can be changed via command line flag)
// then overwrites any config values with command line variables supplied.
func parseConfig() *AppConfig {
	var configFileName string
	flagConfig := AppConfig{}
	flag.StringVar(&configFileName, "config", "config.json", "Specifies the file name where the configuration is kept (JSON).")
	flag.Float64Var(&flagConfig.PollRate, "pollRate", 0, "The per second rate that metrics will be polled.")
	flag.StringVar(&flagConfig.InfluxEndpoint, "influx", "", "Dictates which InfluxDB instance should be used to send metrics.")
	flag.Parse()

	config := defaultConfig
	data, err := ioutil.ReadFile(configFileName)
	if err != nil {
		log.Println("ERROR: Could not read configuration file.\n" + err.Error())
	} else if err = json.Unmarshal(data, config); err != nil {
		log.Println("ERROR: Could not parse configuration file.\n" + err.Error())
	}

	if flagConfig.PollRate != 0 {
		config.PollRate = flagConfig.PollRate
	}

	if flagConfig.InfluxEndpoint != "" {
		config.InfluxEndpoint = flagConfig.InfluxEndpoint
	}

	return config
}

type AppConfig struct {
	PollRate       float64 `json:"pollRate"`
	InfluxEndpoint string  `json:"influxEndpoint"`
}
