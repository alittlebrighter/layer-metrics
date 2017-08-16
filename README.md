# Layer Metrics

## Installation

`go get github.com/alittlebrighter/layer-metrics`

## Description

A service that reads metric data at a specified rate per second (pollRate in config.json) and sends that data to a specified InfluxDB server (influxEndpoint in config.json).  Command line flags override values found in a config file. Run `layer-metrics -h` to see the command line options.

The user interface is an HTTP API served at localhost:8080.  There are two paths /control and /rate.  Any HTTP method directed at /rate returns the currently rate per second in plain text.  A GET to /control will return a brief description of how to send commands in plain text.  A POST to /control with a form value of command=start or command=stop will start and stop metric collection respectively.

When metrics collection has been started it sends random data to the "docbrown" database in the connected InfluxDB instance and records values under the label "fluxCapacitor".

## Note

This service was tested with the latest InfluxDB Docker image pulled 2017-08-16.