package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"time"
)

var neighboursNumber int
var neighbourhoodSize int
var coreAddr string

var workers []string
var neighbourhoods [][]string
var startTime time.Time

var finishedWorkers uint64

type Message struct {
	Op      string
	Content []byte
}

type RegisterMessage struct {
	Addr string `json:"addr"`
}

type NeighbourhoodMessage struct {
	Neighbourhood []string `json:"neighbourhood"`
}

func send(addr string, op string, content []byte) {
	conn, err := net.Dial("unix", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	m := Message{op, content}
	err = encoder.Encode(m)
	if err != nil {
		panic(err)
	}
}

func startWorkers() {

	// create neighbourhoods
	var nh []string
	for i := 0; i < len(workers); i++ {
		nh = append(nh, workers[i])
		if (i+1)%neighbourhoodSize == 0 {
			neighbourhoods = append(neighbourhoods, nh)
			nh = []string{}
		}
	}

	// send workers their neighbourhood
	for i := 0; i < neighboursNumber; i++ {
		for j := 0; j < neighbourhoodSize; j++ {
			nm := NeighbourhoodMessage{neighbourhoods[i]}
			b, err := json.Marshal(nm)
			if err != nil {
				panic(err)
			}
			send(neighbourhoods[i][j], "neighbourhood", b)
		}
	}

	startTime = time.Now()

	// pong workers to start
	for i := 0; i < len(workers); i++ {
		send(workers[i], "pong", nil)
	}
}

func registerWorker(workerAddr string) {
	workers = append(workers, workerAddr)
	if len(workers) == neighboursNumber*neighbourhoodSize {
		startWorkers()
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	var m Message
	err := decoder.Decode(&m)
	if err != nil {
		panic(err)
	}

	switch m.Op {
	case "register":
		var rm RegisterMessage
		err := json.Unmarshal(m.Content, &rm)
		if err != nil {
			panic(err)
		}
		registerWorker(rm.Addr)
	case "finish":
		finishedWorkers++
		if finishedWorkers == uint64(len(workers)) {
			elapsedTime := time.Since(startTime)
			log.Println("Elapsed time: ", elapsedTime.Seconds())
			os.Exit(0)
		}
	}
}

func main() {
	neighboursNumberPtr := flag.Int("nn", 10, "neighbours number")
	neighbourhoodSizePtr := flag.Int("ns", 10, "neighbourhood size")
	coreAddrPtr := flag.String("ca", "/tmp/core.sock", "Core UDS")

	flag.Parse()
	neighboursNumber = *neighboursNumberPtr
	neighbourhoodSize = *neighbourhoodSizePtr
	coreAddr = *coreAddrPtr

	if err := os.RemoveAll(coreAddr); err != nil {
		panic(err)
	}

	l, err := net.Listen("unix", coreAddr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		handleConnection(conn)
	}
}
