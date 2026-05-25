package main

/*

for i in {1..20}
do
  curl http://localhost:9902/order
done

*/

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	circuitOpen bool
	mutex       sync.Mutex
	failures    int
)

func unstableService() error {

	time.Sleep(2 * time.Second)

	if rand.Intn(100) < 70 {
		return fmt.Errorf("service failure")
	}

	return nil
}

func orderHandler(w http.ResponseWriter, r *http.Request) {

	mutex.Lock()

	if circuitOpen {

		mutex.Unlock()

		log.Println("CIRCUIT OPEN - REQUEST BLOCKED")

		http.Error(w, "Service Unavailable", 503)

		return
	}

	mutex.Unlock()

	err := unstableService()

	if err != nil {

		mutex.Lock()

		failures++

		log.Printf("FAILURE COUNT=%d", failures)

		if failures >= 3 {

			circuitOpen = true

			log.Println("CIRCUIT BREAKER OPENED")

			go func() {

				time.Sleep(10 * time.Second)

				mutex.Lock()

				failures = 0
				circuitOpen = false

				mutex.Unlock()

				log.Println("CIRCUIT BREAKER RESET")
			}()
		}

		mutex.Unlock()

		http.Error(w, "Failure", 500)

		return
	}

	log.Println("ORDER SUCCESS")

	fmt.Fprintf(w, "Order Success")
}

func main() {

	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/order", orderHandler)

	log.Println("CIRCUIT BREAKER LAB RUNNING ON PORT 9902")

	http.ListenAndServe(":9902", nil)
}
