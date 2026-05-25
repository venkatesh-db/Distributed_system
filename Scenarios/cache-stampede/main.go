package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type CacheItem struct {
	Value     string
	ExpiresAt time.Time
}

var (
	cache      = map[string]CacheItem{}
	cacheMutex sync.Mutex

	dbHits int
)

func fetchFromDatabase(productID string) string {

	dbHits++

	log.Printf(
		"DB HIT COUNT=%d PRODUCT=%s",
		dbHits,
		productID,
	)

	time.Sleep(3 * time.Second)

	return "PRICE=5000"
}

func getProductPrice(productID string) string {

	cacheMutex.Lock()

	item, exists := cache[productID]

	if exists && time.Now().Before(item.ExpiresAt) {

		cacheMutex.Unlock()

		log.Println("CACHE HIT")

		return item.Value
	}

	cacheMutex.Unlock()

	log.Println("CACHE MISS")

	value := fetchFromDatabase(productID)

	cacheMutex.Lock()

	cache[productID] = CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(5 * time.Second),
	}

	cacheMutex.Unlock()

	return value
}

func productHandler(w http.ResponseWriter, r *http.Request) {

	price := getProductPrice("IPHONE-15")

	fmt.Fprintf(w, price)
}

func trafficExplosion() {

	client := http.Client{}

	for i := 1; i <= 30; i++ {

		go func(requestID int) {

			_, err := client.Get(
				"http://localhost:9903/product",
			)

			if err != nil {

				log.Printf(
					"REQUEST=%d FAILED",
					requestID,
				)

				return
			}

			log.Printf(
				"REQUEST=%d COMPLETED",
				requestID,
			)

		}(i)
	}
}

func metricsPrinter() {

	for {

		time.Sleep(5 * time.Second)

		log.Println("------------------------")
		log.Printf("TOTAL DB HITS=%d", dbHits)
		log.Printf("CACHE SIZE=%d", len(cache))
		log.Println("------------------------")
	}
}

func main() {

	http.HandleFunc("/product", productHandler)

	go metricsPrinter()

	go func() {

		time.Sleep(2 * time.Second)

		log.Println("STARTING TRAFFIC EXPLOSION")

		trafficExplosion()
	}()

	log.Println(
		"CACHE STAMPEDE LAB RUNNING ON PORT 9903",
	)

	err := http.ListenAndServe(":9903", nil)

	if err != nil {
		log.Fatal(err)
	}
}
