package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

const metricName = "fluxCapacitor"

func main() {
	config := parseConfig()

	var module Module = new(Production)

	module.Build(config)

	datastore, err := module.DataStore()
	if err != nil {
		log.Fatalln("Could not connect to datastore! - " + err.Error())
	}

	controller := module.Controller()
	commands, serverExit := controller.Serve("localhost:8080")
	go func() {
		if err := <-serverExit; err != nil {
			log.Fatalln("Controller exited. - " + err.Error())
		}
	}()

	job := module.Job(commands)
	results := job.Run()

	hostname, _ := os.Hostname()
	tags := map[string]string{"host": hostname, "pid": strconv.Itoa(os.Getpid())}
	for result := range results {
		// empty result indicates measurement collection has stopped
		if len(result) == 0 {
			datastore.Flush()
			continue
		}

		datastore.SavePoint(metricName, tags, result, time.Now())
	}

}
