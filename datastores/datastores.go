package datastores

import (
	"log"
	"sync"
	"time"

	idb "github.com/influxdata/influxdb/client/v2"
)

const (
	db       = "docbrown"
	username = ""
	password = ""
)

type DataStore interface {
	SavePoint(string, map[string]string, map[string]interface{}, time.Time) error
	Flush() error
}

type StoreBatch struct {
	Batch    idb.BatchPoints
	Flushing bool
}

type InfluxStore struct {
	client       idb.Client
	expectedFlow float64

	mutex   *sync.RWMutex
	batches []*StoreBatch
}

func NewInfluxStore(host string, expectedFlow float64) (store *InfluxStore, err error) {
	store = new(InfluxStore)
	store.client, err = idb.NewHTTPClient(idb.HTTPConfig{
		Addr:     host,
		Username: username,
		Password: password,
	})
	// make sure the database is there
	_, err = store.client.Query(idb.Query{
		Command:  "CREATE DATABASE " + db,
		Database: "_internal",
	})
	if err != nil {
		return
	}
	store.expectedFlow = expectedFlow

	store.mutex = new(sync.RWMutex)
	store.batches = []*StoreBatch{}
	return
}

// Precision gives the coarsest time unit possible with the given expectedFlow (saves/second)
func (is *InfluxStore) Precision() string {
	switch {
	case is.expectedFlow > 1000:
		return "ns"
	case is.expectedFlow > 1:
		return "ms"
	case is.expectedFlow > 1/60:
		return "s"
	case is.expectedFlow > 1/3600:
		return "m"
	default:
		return "h"
	}
}

// BatchCapacity returns the number of batches expected to be generated in 10 seconds
func (is *InfluxStore) BatchCapacity() int64 {
	return IterationsInSeconds(10, is.expectedFlow)
}

func IterationsInSeconds(seconds int64, ratePerSecond float64) int64 {
	return seconds * int64(time.Second) / int64(1/ratePerSecond*float64(time.Second))
}

// SavePoint collects data points into batches and once a batch has reached the capacity
func (is *InfluxStore) SavePoint(name string, tags map[string]string, fields map[string]interface{}, moment time.Time) (err error) {
	var currentBatch *StoreBatch
	is.mutex.Lock()
	// if there aren't any batches or if the latest batch is being flushed the add a new batch
	if len(is.batches) == 0 || is.batches[len(is.batches)-1].Flushing {
		newBatch := new(StoreBatch)
		newBatch.Batch, err = idb.NewBatchPoints(idb.BatchPointsConfig{
			Database:  db,
			Precision: is.Precision(),
		})
		if err != nil {
			is.mutex.Unlock()
			return
		}

		is.batches = append(is.batches, newBatch)
	}
	currentBatch = is.batches[len(is.batches)-1]
	is.mutex.Unlock()

	var pt *idb.Point
	pt, err = idb.NewPoint(name, tags, fields, moment)
	currentBatch.Batch.AddPoint(pt)

	if int64(len(currentBatch.Batch.Points())) >= is.BatchCapacity() {
		go is.Flush()
	}

	return
}

// Flush is overly complicated due to the fact that writing a batch involves a network call so a batch could
// take longer than expected to perform the Write and we don't want to block new Points from coming in.
func (is *InfluxStore) Flush() (err error) {
	var toRemove int
	is.mutex.RLock()
	for index, storeBatch := range is.batches {
		storeBatch.Flushing = true
		if err = is.client.Write(storeBatch.Batch); err != nil {
			log.Println("Batch failed to write: " + err.Error())
		} else {
			log.Println("Batch successfully written.")
		}
		toRemove = index
	}
	is.mutex.RUnlock()

	is.mutex.Lock()
	if toRemove+1 == len(is.batches) {
		is.batches = []*StoreBatch{}
	} else {
		is.batches = is.batches[toRemove+1:]
	}
	is.mutex.Unlock()
	return
}
