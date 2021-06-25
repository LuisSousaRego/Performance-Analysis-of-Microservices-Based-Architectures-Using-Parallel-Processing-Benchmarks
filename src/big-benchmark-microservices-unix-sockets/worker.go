package main

import (
	"encoding/json"
	"flag"
	"math/rand"
	"net"
	"os"
	"time"
)

var pingLimit int
var pingCounter = 0
var neighbourhood []string
var myAddr string
var coreAddr string

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

type PingMessage struct {
	Addr string `json:"addr"`
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

func register(coreAddr string, workerAddr string) {
	nm := RegisterMessage{workerAddr}
	b, err := json.Marshal(nm)
	if err != nil {
		panic(err)
	}
	send(coreAddr, "register", b)
}

func ping() {
	pingCounter++
	if pingLimit > pingCounter {
		neighbourAddr := getRandomNeighbourAddr(myAddr)
		pm := PingMessage{myAddr}
		b, err := json.Marshal(pm)
		if err != nil {
			panic(err)
		}
		send(neighbourAddr, "ping", b)
	} else {
		send(coreAddr, "finish", nil)
	}
}

func pong(addr string) {
	send(addr, "pong", nil)
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
	case "neighbourhood":
		var nm NeighbourhoodMessage
		err := json.Unmarshal(m.Content, &nm)
		if err != nil {
			panic(err)
		}
		neighbourhood = nm.Neighbourhood
	case "ping":
		var pm PingMessage
		err := json.Unmarshal(m.Content, &pm)
		if err != nil {
			panic(err)
		}
		pong(pm.Addr)
	case "pong":
		ping()
	}
}

func getRandomNeighbourAddr(myAddr string) string {
	// generating a random Addr different from itself
	neighbourAddr := myAddr
	for neighbourAddr != myAddr {
		rand.Seed(time.Now().Unix())
		neighbourAddr = neighbourhood[rand.Intn(len(neighbourhood))]
	}
	return neighbourAddr
}

func main() {
	coreAddrPtr := flag.String("ca", "/tmp/core.sock", "Core UDS")
	workerIdPtr := flag.String("id", "", "Worker id")
	pingLimitPtr := flag.Int("pl", 250000, "ping limit")
	flag.Parse()

	coreAddr = *coreAddrPtr
	pingLimit = *pingLimitPtr
	myAddr = "/tmp/worker" + *workerIdPtr + ".sock"

	if err := os.RemoveAll(myAddr); err != nil {
		panic(err)
	}

	l, err := net.Listen("unix", myAddr)
	if err != nil {
		panic(err)
	}

	register(coreAddr, myAddr)

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		handleConnection(conn)
	}

}
