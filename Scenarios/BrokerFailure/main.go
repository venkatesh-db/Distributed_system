package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

type Job struct {
	ID string
}

var (
	queue           = make(chan Job, 200)
	totalRequests   uint64
	failedRequests  uint64
	successRequests uint64
	workerFailures  uint64
	paymentFailures uint64
)

func unstablePaymentService(jobID string) error {

	latency := rand.Intn(5) + 1

	time.Sleep(time.Duration(latency) * time.Second)

	failure := rand.Intn(100)

	if failure < 60 {

		atomic.AddUint64(&paymentFailures, 1)

		return fmt.Errorf("payment service failure")
	}

	return nil
}

func notificationWorker(workerID int) {

	for job := range queue {

		log.Printf(
			"WORKER=%d PROCESSING=%s",
			workerID,
			job.ID,
		)

		time.Sleep(4 * time.Second)

		failure := rand.Intn(100)

		if failure < 30 {

			atomic.AddUint64(&workerFailures, 1)

			log.Printf(
				"WORKER=%d FAILED=%s",
				workerID,
				job.ID,
			)

			continue
		}

		log.Printf(
			"WORKER=%d COMPLETED=%s",
			workerID,
			job.ID,
		)
	}
}

func orderHandler(w http.ResponseWriter, r *http.Request) {

	requestID := atomic.AddUint64(
		&totalRequests,
		1,
	)

	orderID := fmt.Sprintf(
		"ORD-%d",
		time.Now().UnixNano(),
	)

	log.Printf(
		"REQUEST=%d ORDER=%s RECEIVED",
		requestID,
		orderID,
	)

	var err error

	for retry := 1; retry <= 3; retry++ {

		log.Printf(
			"ORDER=%s PAYMENT RETRY=%d",
			orderID,
			retry,
		)

		err = unstablePaymentService(orderID)

		if err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}

	if err != nil {

		atomic.AddUint64(&failedRequests, 1)

		log.Printf(
			"ORDER=%s FINAL FAILURE",
			orderID,
		)

		http.Error(
			w,
			"Payment Failed",
			http.StatusInternalServerError,
		)

		return
	}

	atomic.AddUint64(&successRequests, 1)

	queue <- Job{
		ID: orderID,
	}

	log.Printf(
		"ORDER=%s SUCCESS QUEUED",
		orderID,
	)

	fmt.Fprintf(
		w,
		"Order Success: %s",
		orderID,
	)
}

func trafficExplosion() {

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	for i := 1; i <= 50; i++ {

		go func(requestNo int) {

			resp, err := client.Get(
				"http://localhost:9906/order",
			)

			if err != nil {

				log.Printf(
					"CLIENT=%d FAILED",
					requestNo,
				)

				return
			}

			log.Printf(
				"CLIENT=%d STATUS=%d",
				requestNo,
				resp.StatusCode,
			)

			resp.Body.Close()

		}(i)
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
			"PAYMENT_FAILURES=%d",
			paymentFailures,
		)

		log.Printf(
			"WORKER_FAILURES=%d",
			workerFailures,
		)

		log.Printf(
			"QUEUE_SIZE=%d",
			len(queue),
		)

		log.Println("===================================")
	}
}

func chaosMonkey() {

	for {

		time.Sleep(15 * time.Second)

		log.Println(
			"CHAOS MONKEY INJECTING LATENCY",
		)

		time.Sleep(5 * time.Second)

		log.Println(
			"CHAOS MONKEY RECOVERED",
		)
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	go notificationWorker(1)
	go notificationWorker(2)

	go metricsPrinter()

	go chaosMonkey()

	http.HandleFunc("/order", orderHandler)

	go func() {

		time.Sleep(3 * time.Second)

		log.Println(
			"STARTING TRAFFIC EXPLOSION",
		)

		trafficExplosion()
	}()

	log.Println(
		"BROKEN FAILURE LAB RUNNING ON PORT 9906",
	)

	err := http.ListenAndServe(":9906", nil)

	if err != nil {
		log.Fatal(err)
	}
}
