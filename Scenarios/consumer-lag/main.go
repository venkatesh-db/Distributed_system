package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

type Event struct {
	ID string
}

var queue = make(chan Event, 100)

func producer() {

	for {

		event := Event{
			ID: fmt.Sprintf("EVENT-%d", rand.Intn(100000)),
		}

		queue <- event

		log.Printf("PRODUCED=%s QUEUE_SIZE=%d", event.ID, len(queue))

		time.Sleep(200 * time.Millisecond)
	}
}

func consumer(workerID int) {

	for event := range queue {

		log.Printf(
			"WORKER=%d CONSUMING=%s",
			workerID,
			event.ID,
		)

		time.Sleep(3 * time.Second)

		log.Printf(
			"WORKER=%d COMPLETED=%s",
			workerID,
			event.ID,
		)
	}
}

func metrics() {

	for {

		time.Sleep(5 * time.Second)

		log.Printf("CURRENT_QUEUE_SIZE=%d", len(queue))
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	go producer()

	go consumer(1)
	go consumer(2)

	go metrics()

	select {}
}
