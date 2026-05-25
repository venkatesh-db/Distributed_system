package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type CacheItem struct {
	Value     string
	ExpiresAt time.Time
}

type CacheNode struct {
	Name  string
	Alive bool
	Data  map[string]CacheItem
	Mutex sync.Mutex
}

var cacheNodes = []*CacheNode{
	{
		Name:  "CACHE-NODE-1",
		Alive: true,
		Data:  map[string]CacheItem{},
	},
	{
		Name:  "CACHE-NODE-2",
		Alive: true,
		Data:  map[string]CacheItem{},
	},
	{
		Name:  "CACHE-NODE-3",
		Alive: true,
		Data:  map[string]CacheItem{},
	},
}

func hashKey(key string) int {

	hash := fnv.New32a()

	hash.Write([]byte(key))

	return int(hash.Sum32())
}

func getCacheNode(key string) *CacheNode {

	index := hashKey(key) % len(cacheNodes)

	return cacheNodes[index]
}

func fetchFromDatabase(productID string) string {

	log.Printf(
		"DATABASE HIT FOR PRODUCT=%s",
		productID,
	)

	time.Sleep(2 * time.Second)

	return fmt.Sprintf(
		"PRODUCT=%s PRICE=%d",
		productID,
		rand.Intn(10000),
	)
}

func getProduct(productID string) string {

	node := getCacheNode(productID)

	if !node.Alive {

		log.Printf(
			"%s DOWN - FAILOVER",
			node.Name,
		)

		node = cacheNodes[(hashKey(productID)+1)%len(cacheNodes)]
	}

	node.Mutex.Lock()

	item, exists := node.Data[productID]

	if exists && time.Now().Before(item.ExpiresAt) {

		node.Mutex.Unlock()

		log.Printf(
			"CACHE HIT NODE=%s PRODUCT=%s",
			node.Name,
			productID,
		)

		return item.Value
	}

	node.Mutex.Unlock()

	log.Printf(
		"CACHE MISS NODE=%s PRODUCT=%s",
		node.Name,
		productID,
	)

	value := fetchFromDatabase(productID)

	node.Mutex.Lock()

	node.Data[productID] = CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(10 * time.Second),
	}

	node.Mutex.Unlock()

	return value
}

func productHandler(w http.ResponseWriter, r *http.Request) {

	productID := r.URL.Query().Get("id")

	if productID == "" {
		productID = "IPHONE-15"
	}

	value := getProduct(productID)

	fmt.Fprintf(w, value)
}

func trafficGenerator() {

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	products := []string{
		"IPHONE-15",
		"SAMSUNG-S24",
		"MACBOOK-PRO",
		"ONEPLUS-12",
	}

	for {

		time.Sleep(1 * time.Second)

		go func() {

			product := products[rand.Intn(len(products))]

			_, err := client.Get(
				"http://localhost:9905/product?id=" + product,
			)

			if err != nil {

				log.Println(
					"CLIENT REQUEST FAILED",
				)

				return
			}

		}()
	}
}

func simulateNodeFailure() {

	for {

		time.Sleep(20 * time.Second)

		node := cacheNodes[rand.Intn(len(cacheNodes))]

		node.Alive = false

		log.Printf(
			"%s FAILED",
			node.Name,
		)

		time.Sleep(10 * time.Second)

		node.Alive = true

		log.Printf(
			"%s RECOVERED",
			node.Name,
		)
	}
}

func metricsPrinter() {

	for {

		time.Sleep(5 * time.Second)

		log.Println("-----------------------------")

		for _, node := range cacheNodes {

			log.Printf(
				"NODE=%s ALIVE=%v KEYS=%d",
				node.Name,
				node.Alive,
				len(node.Data),
			)
		}

		log.Println("-----------------------------")
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/product", productHandler)

	go trafficGenerator()

	go simulateNodeFailure()

	go metricsPrinter()

	log.Println(
		"DISTRIBUTED CACHE LAB RUNNING ON PORT 9905",
	)

	err := http.ListenAndServe(":9905", nil)

	if err != nil {
		log.Fatal(err)
	}
}
