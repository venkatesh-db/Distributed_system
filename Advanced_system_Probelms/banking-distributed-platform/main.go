package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Transaction struct {
	ID      string
	UserID  string
	Amount  int
	Status  string
	TraceID string
}

type CacheNode struct {
	Name  string
	Alive bool
	Data  map[string]Transaction
	Mutex sync.Mutex
}

type Region struct {
	Name    string
	Alive   bool
	Latency time.Duration
}

type RaftNode struct {
	ID       string
	IsLeader bool
	Alive    bool
}

var (
	transactionQueue = make(chan Transaction, 500)

	totalTransactions   uint64
	failedTransactions  uint64
	successTransactions uint64

	circuitBreakerOpen bool
	circuitMutex       sync.Mutex
	paymentFailures    int

	cacheNodes = []*CacheNode{
		{
			Name:  "CACHE-1",
			Alive: true,
			Data:  map[string]Transaction{},
		},
		{
			Name:  "CACHE-2",
			Alive: true,
			Data:  map[string]Transaction{},
		},
	}

	regions = map[string]*Region{
		"mumbai": {
			Name:    "Mumbai",
			Alive:   true,
			Latency: 1 * time.Second,
		},
		"hyderabad": {
			Name:    "Hyderabad",
			Alive:   true,
			Latency: 2 * time.Second,
		},
	}

	raftCluster = []RaftNode{
		{
			ID:    "NODE-1",
			Alive: true,
		},
		{
			ID:    "NODE-2",
			Alive: true,
		},
		{
			ID:    "NODE-3",
			Alive: true,
		},
	}

	mutex sync.Mutex
)

func trace(traceID string, message string) {

	log.Printf(
		"TRACE=%s | %s",
		traceID,
		message,
	)
}

func electLeader() {

	leaderIndex := rand.Intn(len(raftCluster))

	for i := range raftCluster {
		raftCluster[i].IsLeader = false
	}

	raftCluster[leaderIndex].IsLeader = true

	log.Printf(
		"RAFT LEADER ELECTED=%s",
		raftCluster[leaderIndex].ID,
	)
}

func raftFailureSimulation() {

	for {

		time.Sleep(20 * time.Second)

		for i := range raftCluster {

			if raftCluster[i].IsLeader {

				log.Printf(
					"RAFT LEADER FAILED=%s",
					raftCluster[i].ID,
				)

				raftCluster[i].Alive = false
				raftCluster[i].IsLeader = false

				electLeader()

				time.Sleep(10 * time.Second)

				raftCluster[i].Alive = true

				log.Printf(
					"RAFT NODE RECOVERED=%s",
					raftCluster[i].ID,
				)

				break
			}
		}
	}
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

func paymentService(traceID string) error {

	trace(traceID, "PAYMENT STARTED")

	time.Sleep(
		time.Duration(rand.Intn(3)+1) * time.Second,
	)

	if rand.Intn(100) < 40 {

		paymentFailures++

		trace(traceID, "PAYMENT FAILED")

		return fmt.Errorf("payment failure")
	}

	trace(traceID, "PAYMENT SUCCESS")

	return nil
}

func circuitBreaker(traceID string) bool {

	circuitMutex.Lock()

	defer circuitMutex.Unlock()

	if circuitBreakerOpen {

		trace(traceID, "CIRCUIT BREAKER OPEN")

		return false
	}

	if paymentFailures >= 3 {

		circuitBreakerOpen = true

		trace(traceID, "CIRCUIT BREAKER ACTIVATED")

		go func() {

			time.Sleep(10 * time.Second)

			circuitMutex.Lock()

			paymentFailures = 0
			circuitBreakerOpen = false

			circuitMutex.Unlock()

			log.Println(
				"CIRCUIT BREAKER RESET",
			)
		}()

		return false
	}

	return true
}

func transactionProcessor(workerID int) {

	for transaction := range transactionQueue {

		log.Printf(
			"WORKER=%d PROCESSING=%s",
			workerID,
			transaction.ID,
		)

		time.Sleep(2 * time.Second)

		log.Printf(
			"WORKER=%d COMPLETED=%s",
			workerID,
			transaction.ID,
		)
	}
}

func replicateTransaction(transaction Transaction) {

	for _, region := range regions {

		if !region.Alive {

			log.Printf(
				"REGION DOWN=%s",
				region.Name,
			)

			continue
		}

		time.Sleep(region.Latency)

		log.Printf(
			"TRANSACTION=%s REPLICATED TO=%s",
			transaction.ID,
			region.Name,
		)
	}
}

func transactionHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	transactionID := fmt.Sprintf(
		"TXN-%d",
		time.Now().UnixNano(),
	)

	traceID := fmt.Sprintf(
		"TRACE-%d",
		time.Now().UnixNano(),
	)

	atomic.AddUint64(
		&totalTransactions,
		1,
	)

	trace(traceID, "TRANSACTION RECEIVED")

	if !circuitBreaker(traceID) {

		http.Error(
			w,
			"Payment Service Unavailable",
			503,
		)

		return
	}

	transaction := Transaction{
		ID:      transactionID,
		UserID:  "USER-101",
		Amount:  rand.Intn(10000),
		Status:  "PENDING",
		TraceID: traceID,
	}

	node := getCacheNode(transaction.ID)

	node.Mutex.Lock()

	node.Data[transaction.ID] = transaction

	node.Mutex.Unlock()

	trace(
		traceID,
		fmt.Sprintf(
			"CACHE STORED NODE=%s",
			node.Name,
		),
	)

	err := paymentService(traceID)

	if err != nil {

		atomic.AddUint64(
			&failedTransactions,
			1,
		)

		http.Error(
			w,
			"Transaction Failed",
			500,
		)

		return
	}

	transaction.Status = "SUCCESS"

	transactionQueue <- transaction

	go replicateTransaction(transaction)

	atomic.AddUint64(
		&successTransactions,
		1,
	)

	trace(traceID, "TRANSACTION SUCCESS")

	fmt.Fprintf(
		w,
		"Transaction Success=%s",
		transaction.ID,
	)
}

func trafficExplosion() {

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	for wave := 1; wave <= 5; wave++ {

		log.Printf(
			"TRAFFIC WAVE=%d",
			wave,
		)

		for i := 1; i <= 50; i++ {

			go func(clientID int) {

				resp, err := client.Get(
					"http://localhost:9910/transaction",
				)

				if err != nil {

					log.Printf(
						"CLIENT=%d FAILED",
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

func regionFailureSimulation() {

	for {

		time.Sleep(25 * time.Second)

		regions["mumbai"].Alive = false

		log.Println(
			"MUMBAI REGION FAILED",
		)

		time.Sleep(10 * time.Second)

		regions["mumbai"].Alive = true

		log.Println(
			"MUMBAI REGION RECOVERED",
		)
	}
}

func metricsPrinter() {

	for {

		time.Sleep(5 * time.Second)

		log.Println("================================")

		log.Printf(
			"TOTAL_TRANSACTIONS=%d",
			totalTransactions,
		)

		log.Printf(
			"SUCCESS_TRANSACTIONS=%d",
			successTransactions,
		)

		log.Printf(
			"FAILED_TRANSACTIONS=%d",
			failedTransactions,
		)

		log.Printf(
			"QUEUE_SIZE=%d",
			len(transactionQueue),
		)

		log.Printf(
			"CIRCUIT_BREAKER_OPEN=%v",
			circuitBreakerOpen,
		)

		log.Println("================================")
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	electLeader()

	go transactionProcessor(1)
	go transactionProcessor(2)

	go metricsPrinter()

	go trafficExplosion()

	go regionFailureSimulation()

	go raftFailureSimulation()

	http.HandleFunc(
		"/transaction",
		transactionHandler,
	)

	log.Println(
		"BANKING DISTRIBUTED PLATFORM RUNNING ON PORT 9910",
	)

	err := http.ListenAndServe(
		":9910",
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}
}
