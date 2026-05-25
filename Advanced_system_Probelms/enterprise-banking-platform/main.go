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

type RaftNode struct {
	ID       string
	IsLeader bool
	Alive    bool
}

var (

	// ASYNC QUEUE

	transactionQueue = make(chan Transaction, 500)

	// METRICS

	totalRequests   uint64
	successRequests uint64
	failedRequests  uint64

	// CIRCUIT BREAKER

	circuitOpen    bool
	circuitMutex   sync.Mutex
	paymentFailure int

	// DISTRIBUTED CACHE

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

	// RAFT CLUSTER

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
)

func trace(traceID string, message string) {

	log.Printf(
		"TRACE=%s | %s",
		traceID,
		message,
	)
}

//////////////////////////////////////////////////////
// PROBLEM 1
// NETWORK PARTITION + LEADER FAILURE
//////////////////////////////////////////////////////

func electLeader() {

	leaderIndex := rand.Intn(len(raftCluster))

	for i := range raftCluster {
		raftCluster[i].IsLeader = false
	}

	raftCluster[leaderIndex].IsLeader = true

	log.Printf(
		"NEW RAFT LEADER=%s",
		raftCluster[leaderIndex].ID,
	)
}

func raftFailureSimulation() {

	for {

		time.Sleep(20 * time.Second)

		for i := range raftCluster {

			if raftCluster[i].IsLeader {

				log.Printf(
					"LEADER NODE FAILED=%s",
					raftCluster[i].ID,
				)

				raftCluster[i].Alive = false
				raftCluster[i].IsLeader = false

				log.Println(
					"NETWORK PARTITION DETECTED",
				)

				time.Sleep(3 * time.Second)

				electLeader()

				time.Sleep(10 * time.Second)

				raftCluster[i].Alive = true

				log.Printf(
					"NODE RECOVERED=%s",
					raftCluster[i].ID,
				)

				break
			}
		}
	}
}

//////////////////////////////////////////////////////
// PROBLEM 2
// TRAFFIC EXPLOSION + RETRY STORM
//////////////////////////////////////////////////////

func paymentService(traceID string) error {

	trace(traceID, "PAYMENT STARTED")

	time.Sleep(
		time.Duration(rand.Intn(3)+1) * time.Second,
	)

	if rand.Intn(100) < 50 {

		paymentFailure++

		trace(traceID, "PAYMENT FAILURE")

		return fmt.Errorf("payment failed")
	}

	trace(traceID, "PAYMENT SUCCESS")

	return nil
}

func circuitBreaker(traceID string) bool {

	circuitMutex.Lock()

	defer circuitMutex.Unlock()

	if circuitOpen {

		trace(traceID, "CIRCUIT BREAKER OPEN")

		return false
	}

	if paymentFailure >= 3 {

		circuitOpen = true

		trace(traceID, "CIRCUIT BREAKER ACTIVATED")

		go func() {

			time.Sleep(10 * time.Second)

			circuitMutex.Lock()

			paymentFailure = 0
			circuitOpen = false

			circuitMutex.Unlock()

			log.Println(
				"CIRCUIT BREAKER RESET",
			)
		}()

		return false
	}

	return true
}

//////////////////////////////////////////////////////
// DISTRIBUTED CACHE
//////////////////////////////////////////////////////

func hashKey(key string) int {

	hash := fnv.New32a()

	hash.Write([]byte(key))

	return int(hash.Sum32())
}

func getCacheNode(key string) *CacheNode {

	index := hashKey(key) % len(cacheNodes)

	return cacheNodes[index]
}

//////////////////////////////////////////////////////
// ASYNC QUEUE WORKERS
//////////////////////////////////////////////////////

func transactionWorker(workerID int) {

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

//////////////////////////////////////////////////////
// MAIN TRANSACTION FLOW
//////////////////////////////////////////////////////

func transactionHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	requestID := atomic.AddUint64(
		&totalRequests,
		1,
	)

	traceID := fmt.Sprintf(
		"TRACE-%d",
		requestID,
	)

	transactionID := fmt.Sprintf(
		"TXN-%d",
		time.Now().UnixNano(),
	)

	trace(traceID, "TRANSACTION RECEIVED")

	// CIRCUIT BREAKER

	if !circuitBreaker(traceID) {

		http.Error(
			w,
			"Payment Service Unavailable",
			503,
		)

		return
	}

	// CACHE STORAGE

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

	// RETRY STORM

	var err error

	for retry := 1; retry <= 3; retry++ {

		trace(
			traceID,
			fmt.Sprintf(
				"PAYMENT RETRY=%d",
				retry,
			),
		)

		err = paymentService(traceID)

		if err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}

	if err != nil {

		atomic.AddUint64(
			&failedRequests,
			1,
		)

		trace(traceID, "FINAL FAILURE")

		http.Error(
			w,
			"Transaction Failed",
			500,
		)

		return
	}

	transaction.Status = "SUCCESS"

	transactionQueue <- transaction

	atomic.AddUint64(
		&successRequests,
		1,
	)

	trace(traceID, "TRANSACTION SUCCESS")

	fmt.Fprintf(
		w,
		"Transaction Success=%s",
		transaction.ID,
	)
}

//////////////////////////////////////////////////////
// TRAFFIC EXPLOSION
//////////////////////////////////////////////////////

func trafficExplosion() {

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	for wave := 1; wave <= 5; wave++ {

		log.Printf(
			"TRAFFIC WAVE=%d",
			wave,
		)

		for i := 1; i <= 100; i++ {

			go func(clientID int) {

				resp, err := client.Get(
					"http://localhost:9920/transaction",
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

//////////////////////////////////////////////////////
// METRICS
//////////////////////////////////////////////////////

func metricsPrinter() {

	for {

		time.Sleep(5 * time.Second)

		log.Println("================================")

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
			len(transactionQueue),
		)

		log.Printf(
			"CIRCUIT_OPEN=%v",
			circuitOpen,
		)

		log.Println("================================")
	}
}

//////////////////////////////////////////////////////
// MAIN
//////////////////////////////////////////////////////

func main() {

	rand.Seed(time.Now().UnixNano())

	electLeader()

	go transactionWorker(1)
	go transactionWorker(2)

	go metricsPrinter()

	go raftFailureSimulation()

	go func() {

		time.Sleep(3 * time.Second)

		log.Println(
			"STARTING TRAFFIC EXPLOSION",
		)

		trafficExplosion()
	}()

	http.HandleFunc(
		"/transaction",
		transactionHandler,
	)

	log.Println(
		"ENTERPRISE BANKING PLATFORM RUNNING ON PORT 9920",
	)

	err := http.ListenAndServe(
		":9920",
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}
}
