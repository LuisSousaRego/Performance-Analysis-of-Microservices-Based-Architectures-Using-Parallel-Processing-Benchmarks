package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const corePort = "8090"

var workerPort string

const pingLimit = 5000000

var neighbourhood []string

type StartMessage struct {
	Neighbourhood []string `json:"neighbourhood"`
}

func register(port string) {
	_, err := http.Get("http://localhost:" + port + "/register")
	if err != nil {
		panic(err)
	}
}

func finish() {
	_, err := http.Get("http://localhost:" + corePort + "/finish")
	if err != nil {
		panic(err)
	}
}

func getRandomNeighbourPort(myPort string) string {
	// generating a random port different from itself
	targetPort := myPort
	for targetPort != myPort {
		rand.Seed(time.Now().Unix())
		targetPort = neighbourhood[rand.Intn(len(neighbourhood))]
	}
	return targetPort
}

func startHandler(w http.ResponseWriter, req *http.Request) {
	var s StartMessage
	err := json.NewDecoder(req.Body).Decode(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	neighbourhood = s.Neighbourhood

	// get own port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	workerPort = strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)

	// ping neighbourhood until reach limit
	go func() {
		var pingCounter int
		for pingCounter < pingLimit {
			ping()
			pingCounter++
		}
		finish()
		os.Exit(0)
	}()
}

func ping() {

	n := getRandomNeighbourPort(workerPort)

	_, err := http.Get("http://localhost:" + n + "/ping")
	if err != nil {
		panic(err)
	}
}

func pingHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong")
}

func main() {

	fmt.Println("Starting worker...")

	go func() {
		time.Sleep(5 * time.Second)
		register(corePort)
	}()

	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/start", startHandler)

	http.ListenAndServe(":0", nil)

}
