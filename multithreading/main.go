package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

const neighboursNumber = 10
const neighbourhoodSize = 10
const pingLimit = 200000

type worker struct {
	id   int
	ping chan int
	pong chan string
}

func getRandomNeighbour(myId int, neighbourhood *[neighbourhoodSize]worker) int {
	// generating a random id different from itself
	targetId := myId
	for targetId != myId {
		rand.Seed(time.Now().Unix())
		targetId = rand.Intn(len(*neighbourhood))
	}
	return targetId
}

func work(w worker, neighbourhood *[neighbourhoodSize]worker, wg *sync.WaitGroup) {
	pingCounter := 0
	for {
		select {
		case <-w.pong:
			if pingCounter == pingLimit {
				wg.Done()
			} else {
				targetId := getRandomNeighbour(w.id, neighbourhood)
				(*neighbourhood)[targetId].ping <- w.id
				pingCounter++
			}
		case pingerId := <-w.ping:
			(*neighbourhood)[pingerId].pong <- "pong"
		}
	}
}

func main() {

	start := time.Now()

	var neighbourhoods [neighboursNumber][neighbourhoodSize]worker
	var wg sync.WaitGroup

	// create neighbourhoods
	for i := 0; i < neighboursNumber; i++ {
		for j := 0; j < neighbourhoodSize; j++ {
			w := worker{id: j, ping: make(chan int, 1), pong: make(chan string, 1)}
			neighbourhoods[i][j] = w
		}
	}

	// initialize workers
	for i := 0; i < len(neighbourhoods); i++ {
		for j := 0; j < len(neighbourhoods[i]); j++ {
			wg.Add(1)
			go work(neighbourhoods[i][j], &neighbourhoods[i], &wg)
		}
	}

	log.Println("neighboursNumber:", neighboursNumber)
	log.Println("neighbourhoodSize:", neighbourhoodSize)
	log.Println("pingLimit:", pingLimit)

	// start all workers
	for i := 0; i < neighboursNumber; i++ {
		for j := 0; j < neighbourhoodSize; j++ {
			neighbourhoods[i][j].pong <- "pong"
		}
	}

	wg.Wait()

	elapsed := time.Since(start)
	log.Println("Elapsed time:", elapsed)
	log.Println("---")

}
