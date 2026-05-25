package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Order struct {
	OrderID string
}

var (
	requestCounter    uint64
	failureCounter    uint64
	circuitOpen       bool
	circuitMutex      sync.Mutex
	notificationQueue = make(chan Order, 100)
)

func traceLog(traceID string, message string) {

	log.Printf(
		"TRACE_ID=%s | %s",
		traceID,
		message,
	)
}

func paymentService(traceID string) error {

	traceLog(traceID, "PAYMENT SERVICE STARTED")

	time.Sleep(2 * time.Second)

	failure := rand.Intn(100)

	if failure < 40 {

		atomic.AddUint64(&failureCounter, 1)

		traceLog(traceID, "PAYMENT SERVICE FAILED")

		return fmt.Errorf("payment failure")
	}

	traceLog(traceID, "PAYMENT SUCCESSFUL")

	return nil
}

func notificationWorker(workerID int) {

	for order := range notificationQueue {

		log.Printf(
			"WORKER-%d PROCESSING NOTIFICATION: %s",
			workerID,
			order.OrderID,
		)

		time.Sleep(3 * time.Second)

		log.Printf(
			"WORKER-%d COMPLETED NOTIFICATION: %s",
			workerID,
			order.OrderID,
		)
	}
}

func circuitBreaker(traceID string) bool {

	circuitMutex.Lock()
	defer circuitMutex.Unlock()

	if circuitOpen {

		traceLog(traceID, "CIRCUIT BREAKER OPEN")

		return false
	}

	failures := atomic.LoadUint64(&failureCounter)

	if failures >= 3 {

		circuitOpen = true

		traceLog(traceID, "CIRCUIT BREAKER TRIGGERED")

		go func() {

			time.Sleep(10 * time.Second)

			circuitMutex.Lock()

			failureCounter = 0
			circuitOpen = false

			circuitMutex.Unlock()

			log.Println("CIRCUIT BREAKER RESET")
		}()

		return false
	}

	return true
}

func orderHandler(w http.ResponseWriter, r *http.Request) {

	requestID := atomic.AddUint64(&requestCounter, 1)

	traceID := fmt.Sprintf("TRACE-%d", requestID)

	traceLog(traceID, "ORDER REQUEST RECEIVED")

	if !circuitBreaker(traceID) {

		http.Error(
			w,
			"Service Temporarily Unavailable",
			http.StatusServiceUnavailable,
		)

		return
	}

	var err error

	for retry := 1; retry <= 3; retry++ {

		traceLog(
			traceID,
			fmt.Sprintf("PAYMENT RETRY ATTEMPT=%d", retry),
		)

		err = paymentService(traceID)

		if err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}

	if err != nil {

		traceLog(traceID, "ORDER FAILED AFTER RETRIES")

		http.Error(
			w,
			"Payment Failed",
			http.StatusInternalServerError,
		)

		return
	}

	order := Order{
		OrderID: fmt.Sprintf("ORD-%d", time.Now().UnixNano()),
	}

	notificationQueue <- order

	traceLog(traceID, "ORDER SUCCESSFULLY CREATED")

	fmt.Fprintf(
		w,
		"Order Created Successfully: %s",
		order.OrderID,
	)
}

func trafficExplosionSimulation() {

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 1; i <= 20; i++ {

		go func(requestNo int) {

			resp, err := client.Get(
				"http://localhost:9800/order",
			)

			if err != nil {

				log.Println(
					"TRAFFIC REQUEST FAILED:",
					requestNo,
				)

				return
			}

			defer resp.Body.Close()

			log.Printf(
				"TRAFFIC REQUEST=%d STATUS=%d",
				requestNo,
				resp.StatusCode,
			)

		}(i)
	}
	
}

func metricsPrinter() {

	for {

		time.Sleep(5 * time.Second)

		log.Println("-------------------------------")
		log.Printf("TOTAL REQUESTS: %d", requestCounter)
		log.Printf("TOTAL FAILURES: %d", failureCounter)
		log.Printf("QUEUE SIZE: %d", len(notificationQueue))
		log.Printf("CIRCUIT OPEN: %v", circuitOpen)
		log.Println("-------------------------------")
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	go notificationWorker(1)
	go notificationWorker(2)
	go metricsPrinter()

	http.HandleFunc("/order", orderHandler)

	go func() {

		time.Sleep(3 * time.Second)

		log.Println("STARTING TRAFFIC EXPLOSION TEST")

		trafficExplosionSimulation()
	}()

	log.Println("DISTRIBUTED SCENARIO LAB RUNNING ON PORT 9800")

	err := http.ListenAndServe(":9800", nil)

	if err != nil {
		log.Fatal(err)
	}
}
