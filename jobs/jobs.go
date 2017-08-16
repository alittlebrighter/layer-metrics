package jobs

import (
	"log"
	"math/rand"
	"time"

	"github.com/alittlebrighter/layer-metrics/io"
)

type Job interface {
	Run() <-chan map[string]interface{}
}

type FluxCapacitorStatus struct {
	control      <-chan io.Command
	pollInterval time.Duration
}

func NewFluxCapacitorStatus(control <-chan io.Command, pollPerSecond float64) (job *FluxCapacitorStatus) {
	job = new(FluxCapacitorStatus)
	job.control = control

	if pollPerSecond < 0 {
		pollPerSecond = pollPerSecond * (-1)
	}
	job.pollInterval = time.Duration(1 / pollPerSecond * float64(time.Second))
	log.Printf("jobs: Setting poll interval to %v\n", job.pollInterval)
	return
}

// Run starts listening to the control channel.
func (fcs *FluxCapacitorStatus) Run() <-chan map[string]interface{} {
	results := make(chan map[string]interface{})
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	go func() {
		for cmd := range fcs.control {
			if cmd == io.Start {
				break
			}
		}

		jobTicker := Ticker(fcs.pollInterval)
		jobTicker(io.Start)
		log.Println("jobs: " + io.Start.String() + "ed collecting")
		for {
			select {
			// sending unknown just returns the current ticker
			case <-jobTicker(io.Unknown).C:
				results <- map[string]interface{}{
					"charge":       random.Int31(),
					"timeVariance": random.Float32(),
					"flux":         random.ExpFloat64(),
				}
			case cmd := <-fcs.control:
				jobTicker(cmd)
				log.Println("jobs: " + cmd.String() + "ed collecting")
			}
		}
	}()

	return results
}

// Ticker allows stopping and starting the ticker at will without losing the reference to the current time.Ticker
func Ticker(pollInterval time.Duration) func(io.Command) *time.Ticker {
	var theTicker *time.Ticker
	return func(cmd io.Command) *time.Ticker {
		switch cmd {
		case io.Start:
			theTicker = time.NewTicker(pollInterval)
		case io.Stop:
			theTicker.Stop()
		}
		return theTicker
	}
}
