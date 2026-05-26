package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

//////////////////////////////////////////////////////
// MESSAGE
//////////////////////////////////////////////////////

type Message struct {
	ID      int
	Payload string
}

//////////////////////////////////////////////////////
// RABBITMQ SIMULATION
//////////////////////////////////////////////////////

type RabbitMQ struct {
	Queue []Message
	Mutex sync.Mutex
}

//////////////////////////////////////////////////////
// KAFKA SIMULATION
//////////////////////////////////////////////////////

type Kafka struct {
	CommitLog []Message
	Mutex     sync.Mutex
}

//////////////////////////////////////////////////////
// RABBITMQ PUBLISH
//////////////////////////////////////////////////////

func (r *RabbitMQ) Publish(
	message Message,
) {

	r.Mutex.Lock()

	r.Queue = append(
		r.Queue,
		message,
	)

	r.Mutex.Unlock()

	log.Printf(
		"[RABBITMQ] MESSAGE PUBLISHED ID=%d",
		message.ID,
	)
}

//////////////////////////////////////////////////////
// RABBITMQ CONSUMER
//////////////////////////////////////////////////////

func (r *RabbitMQ) Consume() {

	for {

		time.Sleep(2 * time.Second)

		r.Mutex.Lock()

		if len(r.Queue) == 0 {

			r.Mutex.Unlock()

			continue
		}

		message := r.Queue[0]

		r.Queue = r.Queue[1:]

		r.Mutex.Unlock()

		log.Printf(
			"[RABBITMQ] PROCESSING ID=%d",
			message.ID,
		)

		//////////////////////////////////////////////////
		// RANDOM CONSUMER FAILURE
		//////////////////////////////////////////////////

		if rand.Intn(100) < 40 {

			log.Printf(
				"[RABBITMQ] CONSUMER CRASHED ID=%d",
				message.ID,
			)

			//////////////////////////////////////////////////
			// REQUEUE MESSAGE
			//////////////////////////////////////////////////

			r.Publish(message)

			continue
		}

		//////////////////////////////////////////////////
		// ACK SUCCESS
		//////////////////////////////////////////////////

		log.Printf(
			"[RABBITMQ] ACK SUCCESS ID=%d",
			message.ID,
		)
	}
}

//////////////////////////////////////////////////////
// KAFKA PRODUCER
//////////////////////////////////////////////////////

func (k *Kafka) Publish(
	message Message,
) {

	k.Mutex.Lock()

	k.CommitLog = append(
		k.CommitLog,
		message,
	)

	k.Mutex.Unlock()

	log.Printf(
		"[KAFKA] MESSAGE APPENDED ID=%d",
		message.ID,
	)
}

//////////////////////////////////////////////////////
// KAFKA CONSUMER
//////////////////////////////////////////////////////

func (k *Kafka) Consume() {

	offset := 0

	for {

		time.Sleep(2 * time.Second)

		k.Mutex.Lock()

		if offset >= len(k.CommitLog) {

			k.Mutex.Unlock()

			continue
		}

		message := k.CommitLog[offset]

		k.Mutex.Unlock()

		log.Printf(
			"[KAFKA] READING OFFSET=%d ID=%d",
			offset,
			message.ID,
		)

		//////////////////////////////////////////////////
		// RANDOM BROKER FAILURE
		//////////////////////////////////////////////////

		if rand.Intn(100) < 40 {

			log.Printf(
				"[KAFKA] BROKER FAILURE ID=%d",
				message.ID,
			)

			//////////////////////////////////////////////////
			// MESSAGE STILL EXISTS IN LOG
			//////////////////////////////////////////////////

			log.Printf(
				"[KAFKA] REPLAYING MESSAGE ID=%d",
				message.ID,
			)

			continue
		}

		log.Printf(
			"[KAFKA] OFFSET COMMITTED ID=%d",
			message.ID,
		)

		offset++
	}
}

//////////////////////////////////////////////////////
// PRODUCER
//////////////////////////////////////////////////////

func producer(
	rabbit *RabbitMQ,
	kafka *Kafka,
) {

	messageID := 1

	for {

		message := Message{
			ID: messageID,
			Payload: fmt.Sprintf(
				"ORDER-%d",
				messageID,
			),
		}

		rabbit.Publish(message)

		kafka.Publish(message)

		messageID++

		time.Sleep(1 * time.Second)
	}
}

//////////////////////////////////////////////////////
// MAIN
//////////////////////////////////////////////////////

func main() {

	rand.Seed(time.Now().UnixNano())

	rabbit := &RabbitMQ{}

	kafka := &Kafka{}

	//////////////////////////////////////////////////
	// START CONSUMERS
	//////////////////////////////////////////////////

	go rabbit.Consume()

	go kafka.Consume()

	//////////////////////////////////////////////////
	// START PRODUCER
	//////////////////////////////////////////////////

	go producer(
		rabbit,
		kafka,
	)

	select {}
}
