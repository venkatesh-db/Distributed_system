package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"sync"
	"time"
)

//////////////////////////////////////////////////////
// ORDER MODEL
//////////////////////////////////////////////////////

type Order struct {
	OrderID string
	UserID  string
	Amount  int
}

//////////////////////////////////////////////////////
// SHARD
//////////////////////////////////////////////////////

type Shard struct {
	ID     string
	Orders map[string]Order
	Mutex  sync.Mutex
}

//////////////////////////////////////////////////////
// SHARD CLUSTER
//////////////////////////////////////////////////////

var shards []*Shard

//////////////////////////////////////////////////////
// CREATE SHARDS
//////////////////////////////////////////////////////

func createShards() {

	for i := 1; i <= 3; i++ {

		shard := &Shard{
			ID: fmt.Sprintf(
				"SHARD-%d",
				i,
			),
			Orders: make(map[string]Order),
		}

		shards = append(
			shards,
			shard,
		)
	}

	log.Println(
		"SHARD CLUSTER CREATED",
	)
}

//////////////////////////////////////////////////////
// HASH FUNCTION
//////////////////////////////////////////////////////

func hashKey(
	key string,
) uint32 {

	hash := fnv.New32a()

	hash.Write([]byte(key))

	return hash.Sum32()
}

//////////////////////////////////////////////////////
// SHARD ROUTER
//////////////////////////////////////////////////////

func getShard(
	userID string,
) *Shard {

	hash := hashKey(userID)

	index := int(hash) % len(shards)

	return shards[index]
}

//////////////////////////////////////////////////////
// STORE ORDER
//////////////////////////////////////////////////////

func storeOrder(
	order Order,
) {

	shard := getShard(order.UserID)

	shard.Mutex.Lock()

	shard.Orders[order.OrderID] = order

	shard.Mutex.Unlock()

	log.Printf(
		"ORDER=%s USER=%s STORED IN %s",
		order.OrderID,
		order.UserID,
		shard.ID,
	)
}

//////////////////////////////////////////////////////
// CLIENT TRAFFIC SIMULATION
//////////////////////////////////////////////////////

func clientTraffic(
	clientID int,
) {

	for {

		order := Order{
			OrderID: fmt.Sprintf(
				"ORD-%d",
				time.Now().UnixNano(),
			),

			UserID: fmt.Sprintf(
				"USER-%d",
				rand.Intn(100),
			),

			Amount: rand.Intn(10000),
		}

		storeOrder(order)

		time.Sleep(
			time.Duration(rand.Intn(2)+1) *
				time.Second,
		)
	}
}

//////////////////////////////////////////////////////
// SHARD STATUS
//////////////////////////////////////////////////////

func shardStatus() {

	for {

		time.Sleep(5 * time.Second)

		fmt.Println()
		fmt.Println("================================")

		totalOrders := 0

		for _, shard := range shards {

			shard.Mutex.Lock()

			orderCount := len(shard.Orders)

			totalOrders += orderCount

			fmt.Printf(
				"%s TOTAL_ORDERS=%d\n",
				shard.ID,
				orderCount,
			)

			shard.Mutex.Unlock()
		}

		fmt.Println("--------------------------------")

		fmt.Printf(
			"CLUSTER TOTAL ORDERS=%d\n",
			totalOrders,
		)

		fmt.Println("================================")
		fmt.Println()
	}
}

//////////////////////////////////////////////////////
// HOT SHARD SIMULATION
//////////////////////////////////////////////////////

func hotShardTraffic() {

	for {

		time.Sleep(10 * time.Second)

		log.Println(
			"HOT USER TRAFFIC STARTED",
		)

		for i := 1; i <= 20; i++ {

			order := Order{
				OrderID: fmt.Sprintf(
					"HOT-ORD-%d",
					time.Now().UnixNano(),
				),

				//////////////////////////////////////////////////
				// SAME USER
				//////////////////////////////////////////////////

				UserID: "VIP-USER",

				Amount: rand.Intn(50000),
			}

			storeOrder(order)
		}
	}
}

//////////////////////////////////////////////////////
// MAIN
//////////////////////////////////////////////////////

func main() {

	rand.Seed(time.Now().UnixNano())

	createShards()

	//////////////////////////////////////////////////
	// CLIENTS
	//////////////////////////////////////////////////

	for i := 1; i <= 10; i++ {

		go clientTraffic(i)
	}

	go shardStatus()

	go hotShardTraffic()

	select {}
}
