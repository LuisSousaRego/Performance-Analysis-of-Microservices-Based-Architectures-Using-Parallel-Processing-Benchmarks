package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const neighboursNumber = 1
const neighbourhoodSize = 2

var workers []string
var neighbourhoods [neighboursNumber][neighbourhoodSize]string
var mutex = &sync.Mutex{}

var finishedWorkers uint64

type RegisterMessage struct {
	Port string `json:"port"`
}

type StartMessage struct {
	Neighbourhood [neighbourhoodSize]string `json:"neighbourhood"`
}

func startWorkers() {

	// create neighbourhoods
	for i := 0; i < neighboursNumber; i++ {
		copy(neighbourhoods[i][:], workers[i*neighbourhoodSize:(i+1)*neighbourhoodSize])
	}

	// start workers
	for i := 0; i < neighboursNumber; i++ {
		for j := 0; j < neighbourhoodSize; j++ {
			url := "http://localhost:" + neighbourhoods[i][j] + "/start"
			msg := StartMessage{neighbourhoods[i]}
			neighbourhoodString, mErr := json.Marshal(msg)
			if mErr != nil {
				panic(mErr)
			}
			jsonStr := []byte(neighbourhoodString)
			res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
			if err != nil {
				panic(err)
			}
			defer res.Body.Close()
		}
	}
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	var r RegisterMessage
	err := json.NewDecoder(req.Body).Decode(&r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	workerPort := r.Port

	mutex.Lock()
	workers = append(workers, workerPort)
	if len(workers) == neighboursNumber*neighbourhoodSize {
		go startWorkers()
	}
	mutex.Unlock()

	fmt.Fprintf(w, "ok")
}

func finishHandler(w http.ResponseWriter, req *http.Request) {
	atomic.AddUint64(&finishedWorkers, 1)
	fmt.Fprintf(w, "ok")
	if int(finishedWorkers) == len(workers) {
		go func() {
			log.Println("exiting")
			time.Sleep(time.Second)
			os.Exit(0)
		}()
	}
}

func main() {

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/finish", finishHandler)

	http.ListenAndServe(":8090", nil)
}
