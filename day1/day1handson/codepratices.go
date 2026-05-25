package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Order struct {
	ID int
}

type Result struct {
	OrderID int
	Status  string
}

var (
	totalProcessed uint64
	totalFailed    uint64
)

func paymentWorker(
	ctx context.Context,
	workerID int,
	orderQueue <-chan Order,
	resultQueue chan<- Result,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	for {

		select {

		case <-ctx.Done():

			log.Printf(
				"PAYMENT WORKER=%d SHUTDOWN",
				workerID,
			)

			return

		case order := <-orderQueue:

			log.Printf(
				"PAYMENT WORKER=%d PROCESSING ORDER=%d",
				workerID,
				order.ID,
			)

			time.Sleep(
				time.Duration(rand.Intn(3)+1) * time.Second,
			)

			if rand.Intn(100) < 20 {

				atomic.AddUint64(
					&totalFailed,
					1,
				)

				log.Printf(
					"PAYMENT FAILED ORDER=%d",
					order.ID,
				)

				resultQueue <- Result{
					OrderID: order.ID,
					Status:  "PAYMENT_FAILED",
				}

				continue
			}

			resultQueue <- Result{
				OrderID: order.ID,
				Status:  "PAYMENT_SUCCESS",
			}
		}
	}
}

func inventoryWorker(
	ctx context.Context,
	workerID int,
	resultQueue <-chan Result,
	notificationQueue chan<- string,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	for {

		select {

		case <-ctx.Done():

			log.Printf(
				"INVENTORY WORKER=%d SHUTDOWN",
				workerID,
			)

			return

		case result := <-resultQueue:

			if result.Status != "PAYMENT_SUCCESS" {

				continue
			}

			log.Printf(
				"INVENTORY WORKER=%d UPDATING ORDER=%d",
				workerID,
				result.OrderID,
			)

			time.Sleep(2 * time.Second)

			notificationQueue <- fmt.Sprintf(
				"ORDER=%d INVENTORY UPDATED",
				result.OrderID,
			)
		}
	}
}

func notificationWorker(
	ctx context.Context,
	workerID int,
	notificationQueue <-chan string,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	for {

		select {

		case <-ctx.Done():

			log.Printf(
				"NOTIFICATION WORKER=%d SHUTDOWN",
				workerID,
			)

			return

		case message := <-notificationQueue:

			log.Printf(
				"NOTIFICATION WORKER=%d MESSAGE=%s",
				workerID,
				message,
			)

			time.Sleep(1 * time.Second)

			atomic.AddUint64(
				&totalProcessed,
				1,
			)
		}
	}
}

func metricsPrinter() {

	for {

		time.Sleep(5 * time.Second)

		log.Println("================================")

		log.Printf(
			"TOTAL_PROCESSED=%d",
			totalProcessed,
		)

		log.Printf(
			"TOTAL_FAILED=%d",
			totalFailed,
		)

		log.Println("================================")
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	defer cancel()

	orderQueue := make(chan Order, 100)

	resultQueue := make(chan Result, 100)

	notificationQueue := make(chan string, 100)

	var wg sync.WaitGroup

	//////////////////////////////////////////////////
	// PAYMENT WORKERS
	//////////////////////////////////////////////////

	for i := 1; i <= 5; i++ {

		wg.Add(1)

		go paymentWorker(
			ctx,
			i,
			orderQueue,
			resultQueue,
			&wg,
		)
	}

	//////////////////////////////////////////////////
	// INVENTORY WORKERS
	//////////////////////////////////////////////////

	for i := 1; i <= 3; i++ {

		wg.Add(1)

		go inventoryWorker(
			ctx,
			i,
			resultQueue,
			notificationQueue,
			&wg,
		)
	}

	//////////////////////////////////////////////////
	// NOTIFICATION WORKERS
	//////////////////////////////////////////////////

	for i := 1; i <= 2; i++ {

		wg.Add(1)

		go notificationWorker(
			ctx,
			i,
			notificationQueue,
			&wg,
		)
	}

	go metricsPrinter()

	//////////////////////////////////////////////////
	// PRODUCER
	//////////////////////////////////////////////////

	for orderID := 1; orderID <= 50; orderID++ {

		orderQueue <- Order{
			ID: orderID,
		}
	}

	time.Sleep(30 * time.Second)

	cancel()

	wg.Wait()

	log.Println(
		"GRACEFUL SHUTDOWN COMPLETE",
	)
}
