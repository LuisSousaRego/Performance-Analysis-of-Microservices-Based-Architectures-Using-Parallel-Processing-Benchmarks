package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

const neighboursNumber = 1
const neighbourhoodSize = 2

var workers []string
var neighbourhoods [neighboursNumber][neighbourhoodSize]string
var mutex = &sync.Mutex{}

var finishedWorkers uint64

type StartMessage struct {
	Neighbourhood [neighbourhoodSize]string `json:"neighbourhood"`
}

func createWorkers() {
	for i := 0; i < neighboursNumber*neighbourhoodSize; i++ {
		go func(i int) {
			//cmdStr := "./worker > workerLog" + strconv.Itoa(i)
			cmd := exec.Command("./worker")

			outfile, ferr := os.Create("workerLog" + strconv.Itoa(i) + ".log")
			if ferr != nil {
				panic(ferr)
			}
			defer outfile.Close()

			writer := bufio.NewWriter(outfile)
			defer writer.Flush()

			workerOut, _ := cmd.StdoutPipe()
			err := cmd.Start()
			if err != nil {
				panic(err)
			}

			go io.Copy(writer, workerOut)
			cmd.Wait()

		}(i)
	}
}

func initWorkers() {
	for i := 0; i < neighboursNumber; i++ {
		for j := 0; j < neighbourhoodSize; j++ {
			url := "http://localhost:" + neighbourhoods[i][j] + "/start"
			msg := StartMessage{neighbourhoods[i]}
			neighbourhoodString, mErr := json.Marshal(msg)
			if mErr != nil {
				panic(mErr)
			}
			jsonStr := []byte(neighbourhoodString)
			_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
			if err != nil {
				panic(err)
			}
		}
	}
}

func createNighbourhoods() {
	for i := 0; i < neighboursNumber; i++ {
		for j := 0; j < neighbourhoodSize; j++ {
			neighbourhoods[i][j] = workers[0]
			workers = workers[1:]
		}
	}
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	addrSlice := strings.Split(req.RemoteAddr, ":")
	workerPort := addrSlice[len(addrSlice)-1]
	mutex.Lock()
	workers = append(workers, workerPort)
	mutex.Unlock()
	fmt.Println("registered worker on port: ", workerPort)
	if len(workers) == neighboursNumber*neighbourhoodSize {
		createNighbourhoods()
		initWorkers()
	}
}

func finishHandler(w http.ResponseWriter, req *http.Request) {
	atomic.AddUint64(&finishedWorkers, 1)

	if int(finishedWorkers) == len(workers) {
		os.Exit(0)
	}
}

func main() {

	createWorkers()

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/finish", finishHandler)

	http.ListenAndServe(":8090", nil)
}
