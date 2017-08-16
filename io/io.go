package io

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Command dictates the type for different commands the user can send to the system.
type Command uint8

const (
	Start Command = iota
	Stop
	Unknown // Unknown purely exists so that ParseCommand doesn't have to return an error
)

func (c Command) String() string {
	switch c {
	case Start:
		return "START"
	case Stop:
		return "STOP"
	default:
		return "UNKNOWN"
	}
}

// ParseCommand translates a string into a valid Command type.
func ParseCommand(cmdStr string) Command {
	switch strings.ToUpper(cmdStr) {
	case "START":
		return Start
	case "STOP":
		return Stop
	default:
		return Unknown
	}
}

// Controller is responsible for providing an interface that an end-user can interact with.
type Controller interface {
	Serve(string) (<-chan Command, <-chan error)
}

// HTTPController exposes an HTTP API for end-user interaction.
type HTTPController struct {
	commands chan Command
	pollRate float64
}

func NewHTTPController(pollRate float64) (controller *HTTPController) {
	controller = new(HTTPController)
	controller.commands = make(chan Command, 5) // some buffer so http clients don't timeout
	controller.pollRate = pollRate
	return
}

// Serve starts the http server.
func (hc *HTTPController) Serve(serveAt string) (<-chan Command, <-chan error) {
	http.HandleFunc("/control", hc.controlHandler)
	http.HandleFunc("/rate", hc.rate())

	log.Println("io: HTTPController serving at " + serveAt)
	doneServing := make(chan error)
	go func() {
		doneServing <- http.ListenAndServe(serveAt, nil)
	}()
	return hc.commands, doneServing
}

// ControlHandler accepts commands sent via POST form data.  It also, returns usage documentation
func (hc *HTTPController) controlHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	switch strings.ToUpper(r.Method) {
	case http.MethodGet:
		if _, err := fmt.Fprintf(w, "POST command=start to start metric collection and command=stop to stop."); err != nil {
			log.Println("io: Error writing control description: " + err.Error())
			status = http.StatusInternalServerError
		}
		log.Println("io: Showing documentation.")
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			log.Println("io: ERROR - " + err.Error())
			status = http.StatusBadRequest
			break
		}

		command := ParseCommand(r.PostFormValue("command"))
		if command != Start && command != Stop {
			log.Println("io: unknown command received")
			status = http.StatusBadRequest
			break
		}

		hc.commands <- command

		response := "Command " + command.String() + " received."
		log.Println("io: " + response)
		fmt.Fprintf(w, response)
		status = http.StatusAccepted
	}

	w.WriteHeader(status)
}

func (hc *HTTPController) rate() func(http.ResponseWriter, *http.Request) {
	pollPerSecond := fmt.Sprintf("%f/second", hc.pollRate) // precompute the string since it can't be changed mid-session
	return func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		if _, err := fmt.Fprintf(w, pollPerSecond); err != nil {
			log.Println("Error writing rate response: " + err.Error())
			status = http.StatusInternalServerError
			return
		}

		w.WriteHeader(status)
	}
}
