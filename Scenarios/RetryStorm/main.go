package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

var totalRequests uint64
var failedRequests uint64

func paymentHandler(w http.ResponseWriter, r *http.Request) {

	requestID := atomic.AddUint64(&totalRequests, 1)

	log.Printf("REQUEST=%d RECEIVED", requestID)

	time.Sleep(1 * time.Second)

	failure := rand.Intn(100)

	if failure < 60 {

		atomic.AddUint64(&failedRequests, 1)

		log.Printf("REQUEST=%d PAYMENT FAILED", requestID)

		http.Error(w, "Payment Failed", 500)

		return
	}

	log.Printf("REQUEST=%d PAYMENT SUCCESS", requestID)

	fmt.Fprintf(w, "Payment Success")
}

func clientSimulator(clientID int) {

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	for retry := 1; retry <= 3; retry++ {

		resp, err := client.Get("http://localhost:9901/pay")

		if err != nil {

			log.Printf(
				"CLIENT=%d RETRY=%d NETWORK ERROR",
				clientID,
				retry,
			)

			continue
		}

		if resp.StatusCode == 200 {

			log.Printf(
				"CLIENT=%d SUCCESS AFTER RETRY=%d",
				clientID,
				retry,
			)

			resp.Body.Close()

			return
		}

		log.Printf(
			"CLIENT=%d RETRY=%d STATUS=%d",
			clientID,
			retry,
			resp.StatusCode,
		)

		resp.Body.Close()

		time.Sleep(1 * time.Second)
	}

	log.Printf("CLIENT=%d FINAL FAILURE", clientID)
}

func metricsPrinter() {

	for {

		time.Sleep(5 * time.Second)

		log.Println("--------------------------")
		log.Printf("TOTAL REQUESTS=%d", totalRequests)
		log.Printf("FAILED REQUESTS=%d", failedRequests)
		log.Println("--------------------------")
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/pay", paymentHandler)

	go func() {

		time.Sleep(2 * time.Second)

		for i := 1; i <= 20; i++ {

			go clientSimulator(i)
		}
	}()

	go metricsPrinter()

	log.Println("RETRY STORM LAB RUNNING ON PORT 9901")

	http.ListenAndServe(":9901", nil)
}
