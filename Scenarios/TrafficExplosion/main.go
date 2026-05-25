
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

type Request struct {
	ID string
}

var (
	requestQueue    = make(chan Request, 100)
	totalRequests   uint64
	successRequests uint64
	failedRequests  uint64
	activeWorkers   uint64
)

func processRequest(workerID int) {

	for request := range requestQueue {

		atomic.AddUint64(&activeWorkers, 1)

		log.Printf(
			"WORKER=%d PROCESSING=%s",
			workerID,
			request.ID,
		)

		latency := rand.Intn(5) + 1

		time.Sleep(time.Duration(latency) * time.Second)

		if rand.Intn(100) < 30 {

			atomic.AddUint64(&failedRequests, 1)

			log.Printf(
				"WORKER=%d FAILED=%s",
				workerID,
				request.ID,
			)

			atomic.AddUint64(&activeWorkers, ^uint64(0))

			continue
		}

		atomic.AddUint64(&successRequests, 1)

		log.Printf(
			"WORKER=%d COMPLETED=%s",
			workerID,
			request.ID,
		)

		atomic.AddUint64(&activeWorkers, ^uint64(0))
	}
}

func orderHandler(w http.ResponseWriter, r *http.Request) {

	requestID := atomic.AddUint64(
		&totalRequests,
		1,
	)

	orderID := fmt.Sprintf(
		"REQ-%d",
		time.Now().UnixNano(),
	)

	if len(requestQueue) >= 90 {

		log.Printf(
			"REQUEST=%d DROPPED QUEUE FULL",
			requestID,
		)

		http.Error(
			w,
			"Server Busy",
			http.StatusTooManyRequests,
		)

		return
	}

	requestQueue <- Request{
		ID: orderID,
	}

	log.Printf(
		"REQUEST=%d QUEUED=%s",
		requestID,
		orderID,
	)

	fmt.Fprintf(
		w,
		"Request Accepted: %s",
		orderID,
	)
}

func trafficExplosionSimulator() {

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	for wave := 1; wave <= 5; wave++ {

		log.Printf(
			"TRAFFIC WAVE=%d STARTED",
			wave,
		)

		for i := 1; i <= 100; i++ {

			go func(clientID int) {

				resp, err := client.Get(
					"http://localhost:9907/order",
				)

				if err != nil {

					log.Printf(
						"CLIENT=%d NETWORK FAILURE",
						clientID,
					)

					return
				}

				log.Printf(
					"CLIENT=%d STATUS=%d",
					clientID,
					resp.StatusCode,
				)

				resp.Body.Close()

			}(i)
		}

		time.Sleep(5 * time.Second)
	}
}

func metricsPrinter() {

	for {

		time.Sleep(5 * time.Second)

		log.Println("===================================")

		log.Printf(
			"TOTAL_REQUESTS=%d",
			totalRequests,
		)

		log.Printf(
			"SUCCESS_REQUESTS=%d",
			successRequests,
		)

		log.Printf(
			"FAILED_REQUESTS=%d",
			failedRequests,
		)

		log.Printf(
			"QUEUE_SIZE=%d",
			len(requestQueue),
		)

		log.Printf(
			"ACTIVE_WORKERS=%d",
			activeWorkers,
		)

		log.Println("===================================")
	}
}

func autoscaler() {

	for {

		time.Sleep(10 * time.Second)

		queueSize := len(requestQueue)

		if queueSize > 70 {

			log.Println(
				"AUTOSCALER DETECTED HIGH LOAD",
			)

			go processRequest(rand.Intn(1000))

			log.Println(
				"NEW WORKER SPAWNED",
			)
		}
	}
}

func chaosLatencyInjection() {

	for {

		time.Sleep(20 * time.Second)

		log.Println(
			"CHAOS ENGINEERING: LATENCY SPIKE",
		)

		time.Sleep(5 * time.Second)

		log.Println(
			"LATENCY NORMALIZED",
		)
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	for i := 1; i <= 3; i++ {

		go processRequest(i)
	}

	go metricsPrinter()

	go autoscaler()

	go chaosLatencyInjection()

	http.HandleFunc("/order", orderHandler)

	go func() {

		time.Sleep(3 * time.Second)

		log.Println(
			"STARTING TRAFFIC EXPLOSION TEST",
		)

		trafficExplosionSimulator()
	}()

	log.Println(
		"TRAFFIC EXPLOSION LAB RUNNING ON PORT 9907",
	)

	err := http.ListenAndServe(":9907", nil)

	if err != nil {
		log.Fatal(err)
	}
}