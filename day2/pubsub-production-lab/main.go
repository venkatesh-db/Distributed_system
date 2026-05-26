package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

//////////////////////////////////////////////////////
// EVENT MODEL
//////////////////////////////////////////////////////

type OrderEvent struct {
	OrderID string
	UserID  string
	Amount  int
}

//////////////////////////////////////////////////////
// PUB SUB BROKER
//////////////////////////////////////////////////////

type Broker struct {
	subscribers map[string][]chan OrderEvent
	mutex       sync.RWMutex
}

func NewBroker() *Broker {

	return &Broker{
		subscribers: make(map[string][]chan OrderEvent),
	}
}

func (b *Broker) Subscribe(
	topic string,
) chan OrderEvent {

	channel := make(chan OrderEvent, 100)

	b.mutex.Lock()

	b.subscribers[topic] = append(
		b.subscribers[topic],
		channel,
	)

	b.mutex.Unlock()

	return channel
}

func (b *Broker) Publish(
	topic string,
	event OrderEvent,
) {

	b.mutex.RLock()

	subscribers := b.subscribers[topic]

	b.mutex.RUnlock()

	for _, subscriber := range subscribers {

		select {

		case subscriber <- event:

		default:

			log.Printf(
				"BACKPRESSURE DROPPED EVENT=%s",
				event.OrderID,
			)
		}
	}
}

//////////////////////////////////////////////////////
// GLOBAL METRICS
//////////////////////////////////////////////////////

var (
	totalPublished uint64
	totalProcessed uint64
	totalFailed    uint64
	totalRetried   uint64
)

//////////////////////////////////////////////////////
// EMAIL CONSUMER
//////////////////////////////////////////////////////

func emailConsumer(
	workerID int,
	eventChannel chan OrderEvent,
) {

	for event := range eventChannel {

		log.Printf(
			"[EMAIL-%d] RECEIVED=%s",
			workerID,
			event.OrderID,
		)

		//////////////////////////////////////////////////
		// NETWORK LATENCY SIMULATION
		//////////////////////////////////////////////////

		time.Sleep(
			time.Duration(rand.Intn(5)+2) * time.Second,
		)

		//////////////////////////////////////////////////
		// RANDOM FAILURE
		//////////////////////////////////////////////////

		if rand.Intn(100) < 30 {

			log.Printf(
				"[EMAIL-%d] FAILED=%s",
				workerID,
				event.OrderID,
			)

			atomic.AddUint64(
				&totalFailed,
				1,
			)

			//////////////////////////////////////////////////
			// RETRY
			//////////////////////////////////////////////////

			atomic.AddUint64(
				&totalRetried,
				1,
			)

			log.Printf(
				"[EMAIL-%d] RETRY=%s",
				workerID,
				event.OrderID,
			)

			time.Sleep(2 * time.Second)
		}

		log.Printf(
			"[EMAIL-%d] SUCCESS=%s",
			workerID,
			event.OrderID,
		)

		atomic.AddUint64(
			&totalProcessed,
			1,
		)
	}
}

//////////////////////////////////////////////////////
// INVENTORY CONSUMER
//////////////////////////////////////////////////////

func inventoryConsumer(
	workerID int,
	eventChannel chan OrderEvent,
) {

	for event := range eventChannel {

		log.Printf(
			"[INVENTORY-%d] PROCESSING=%s",
			workerID,
			event.OrderID,
		)

		time.Sleep(
			time.Duration(rand.Intn(3)+1) * time.Second,
		)

		log.Printf(
			"[INVENTORY-%d] STOCK UPDATED=%s",
			workerID,
			event.OrderID,
		)

		atomic.AddUint64(
			&totalProcessed,
			1,
		)
	}
}

//////////////////////////////////////////////////////
// ANALYTICS CONSUMER
//////////////////////////////////////////////////////

func analyticsConsumer(
	workerID int,
	eventChannel chan OrderEvent,
) {

	for event := range eventChannel {

		log.Printf(
			"[ANALYTICS-%d] PROCESSING=%s",
			workerID,
			event.OrderID,
		)

		time.Sleep(
			time.Duration(rand.Intn(4)+1) * time.Second,
		)

		log.Printf(
			"[ANALYTICS-%d] EVENT STORED=%s",
			workerID,
			event.OrderID,
		)

		atomic.AddUint64(
			&totalProcessed,
			1,
		)
	}
}

//////////////////////////////////////////////////////
// PRODUCER
//////////////////////////////////////////////////////

func producer(
	broker *Broker,
) {

	for {

		orderID := fmt.Sprintf(
			"ORD-%d",
			time.Now().UnixNano(),
		)

		event := OrderEvent{
			OrderID: orderID,
			UserID:  "USER-101",
			Amount:  rand.Intn(10000),
		}

		log.Printf(
			"[PRODUCER] PUBLISHING=%s",
			event.OrderID,
		)

		broker.Publish(
			"orders",
			event,
		)

		atomic.AddUint64(
			&totalPublished,
			1,
		)

		time.Sleep(1 * time.Second)
	}
}

//////////////////////////////////////////////////////
// METRICS MONITOR
//////////////////////////////////////////////////////

func metricsMonitor() {

	for {

		time.Sleep(5 * time.Second)

		log.Println("================================")

		log.Printf(
			"TOTAL_PUBLISHED=%d",
			totalPublished,
		)

		log.Printf(
			"TOTAL_PROCESSED=%d",
			totalProcessed,
		)

		log.Printf(
			"TOTAL_FAILED=%d",
			totalFailed,
		)

		log.Printf(
			"TOTAL_RETRIED=%d",
			totalRetried,
		)

		log.Println("================================")
	}
}

//////////////////////////////////////////////////////
// MAIN
//////////////////////////////////////////////////////

func main() {

	rand.Seed(time.Now().UnixNano())

	log.Println(
		"PUB SUB PRODUCTION LAB STARTED",
	)

	broker := NewBroker()

	//////////////////////////////////////////////////
	// SUBSCRIBERS
	//////////////////////////////////////////////////

	emailChannel := broker.Subscribe("orders")

	inventoryChannel := broker.Subscribe("orders")

	analyticsChannel := broker.Subscribe("orders")

	//////////////////////////////////////////////////
	// EMAIL WORKERS
	//////////////////////////////////////////////////

	for i := 1; i <= 3; i++ {

		go emailConsumer(
			i,
			emailChannel,
		)
	}

	//////////////////////////////////////////////////
	// INVENTORY WORKERS
	//////////////////////////////////////////////////

	for i := 1; i <= 2; i++ {

		go inventoryConsumer(
			i,
			inventoryChannel,
		)
	}

	//////////////////////////////////////////////////
	// ANALYTICS WORKERS
	//////////////////////////////////////////////////

	for i := 1; i <= 2; i++ {

		go analyticsConsumer(
			i,
			analyticsChannel,
		)
	}

	go producer(broker)

	go metricsMonitor()

	select {}
}
