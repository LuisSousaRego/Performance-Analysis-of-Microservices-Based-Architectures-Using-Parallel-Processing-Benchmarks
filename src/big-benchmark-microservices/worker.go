package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const corePort = "8090"

var pingLimit int
var workerPort string
var neighbourhood []string

type RegisterMessage struct {
	Port string `json:"port"`
}

type StartMessage struct {
	Neighbourhood []string `json:"neighbourhood"`
	PingLimit     int      `json:"pingLimit"`
}

func register(port string) {
	url := "http://localhost:" + port + "/register"
	msg := RegisterMessage{workerPort}
	registerMessageStr, mErr := json.Marshal(msg)
	if mErr != nil {
		panic(mErr)
	}
	jsonStr := []byte(registerMessageStr)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
}

func finish() {
	res, err := http.Get("http://localhost:" + corePort + "/finish")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
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
	pingLimit = s.PingLimit

	// ping neighbourhood until reach limit
	go func() {
		var pingCounter int
		for pingCounter < pingLimit {
			ping()
			pingCounter++
		}
		finish()
		log.Println("exiting")
		os.Exit(0)
	}()
}

func ping() {
	n := getRandomNeighbourPort(workerPort)
	res, err := http.Get("http://localhost:" + n + "/ping")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
}

func pingHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong")
}

func main() {

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	workerPort = strconv.Itoa(l.Addr().(*net.TCPAddr).Port)

	register(corePort)

	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/start", startHandler)

	http.Serve(l, nil)

}
