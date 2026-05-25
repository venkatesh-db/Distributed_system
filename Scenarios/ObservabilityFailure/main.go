package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

type LogEvent struct {
	TraceID string
	Message string
}

var (
	logQueue = make(chan LogEvent, 200)

	totalRequests uint64
	logFailures   uint64
	traceLoss     uint64

	observabilitySystemAlive = true
)

func telemetryCollector() {

	for event := range logQueue {

		if !observabilitySystemAlive {

			atomic.AddUint64(&traceLoss, 1)

			log.Printf(
				"TRACE LOST=%s MESSAGE=%s",
				event.TraceID,
				event.Message,
			)

			continue
		}

		latency := rand.Intn(3) + 1

		time.Sleep(time.Duration(latency) * time.Second)

		if rand.Intn(100) < 30 {

			atomic.AddUint64(&logFailures, 1)

			log.Printf(
				"TELEMETRY BACKEND FAILURE TRACE=%s",
				event.TraceID,
			)

			continue
		}

		log.Printf(
			"TRACE STORED=%s MESSAGE=%s",
			event.TraceID,
			event.Message,
		)
	}
}

func orderHandler(w http.ResponseWriter, r *http.Request) {

	requestID := atomic.AddUint64(
		&totalRequests,
		1,
	)

	traceID := fmt.Sprintf(
		"TRACE-%d",
		requestID,
	)

	log.Printf(
		"REQUEST RECEIVED TRACE=%s",
		traceID,
	)

	if len(logQueue) >= 180 {

		log.Printf(
			"OBSERVABILITY SATURATED TRACE=%s",
			traceID,
		)

		http.Error(
			w,
			"Monitoring Saturated",
			http.StatusServiceUnavailable,
		)

		return
	}

	logQueue <- LogEvent{
		TraceID: traceID,
		Message: "ORDER_CREATED",
	}

	time.Sleep(1 * time.Second)

	fmt.Fprintf(
		w,
		"Order Processed Trace=%s",
		traceID,
	)
}

func trafficExplosion() {

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	for wave := 1; wave <= 5; wave++ {

		log.Printf(
			"TRAFFIC WAVE=%d",
			wave,
		)

		for i := 1; i <= 50; i++ {

			go func(clientID int) {

				resp, err := client.Get(
					"http://localhost:9908/order",
				)

				if err != nil {

					log.Printf(
						"CLIENT=%d FAILED",
						clientID,
					)

					return
				}

				resp.Body.Close()

			}(i)
		}

		time.Sleep(5 * time.Second)
	}
}

func simulateObservabilityFailure() {

	for {

		time.Sleep(15 * time.Second)

		observabilitySystemAlive = false

		log.Println(
			"OBSERVABILITY SYSTEM FAILED",
		)

		time.Sleep(10 * time.Second)

		observabilitySystemAlive = true

		log.Println(
			"OBSERVABILITY SYSTEM RECOVERED",
		)
	}
}

func metricsPrinter() {

	for {

		time.Sleep(5 * time.Second)

		log.Println("================================")

		log.Printf(
			"TOTAL_REQUESTS=%d",
			totalRequests,
		)

		log.Printf(
			"LOG_FAILURES=%d",
			logFailures,
		)

		log.Printf(
			"TRACE_LOSS=%d",
			traceLoss,
		)

		log.Printf(
			"LOG_QUEUE_SIZE=%d",
			len(logQueue),
		)

		log.Printf(
			"OBSERVABILITY_ALIVE=%v",
			observabilitySystemAlive,
		)

		log.Println("================================")
	}
}

func chaosLatencyInjection() {

	for {

		time.Sleep(20 * time.Second)

		log.Println(
			"CHAOS: TELEMETRY LATENCY SPIKE",
		)

		time.Sleep(5 * time.Second)

		log.Println(
			"CHAOS: LATENCY NORMALIZED",
		)
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	go telemetryCollector()

	go metricsPrinter()

	go simulateObservabilityFailure()

	go chaosLatencyInjection()

	http.HandleFunc("/order", orderHandler)

	go func() {

		time.Sleep(3 * time.Second)

		log.Println(
			"STARTING OBSERVABILITY TRAFFIC TEST",
		)

		trafficExplosion()
	}()

	log.Println(
		"OBSERVABILITY FAILURE LAB RUNNING ON PORT 9908",
	)

	err := http.ListenAndServe(":9908", nil)

	if err != nil {
		log.Fatal(err)
	}
}
