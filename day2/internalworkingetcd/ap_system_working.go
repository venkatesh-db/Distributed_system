package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

//////////////////////////////////////////////////////
// NODE
//////////////////////////////////////////////////////

type Node struct {
	ID    string
	Alive bool
	Data  map[string]string
	Mutex sync.Mutex
}

//////////////////////////////////////////////////////
// CLUSTER
//////////////////////////////////////////////////////

var nodes []*Node

//////////////////////////////////////////////////////
// CREATE CLUSTER
//////////////////////////////////////////////////////

func createCluster() {

	for i := 1; i <= 3; i++ {

		node := &Node{
			ID:    fmt.Sprintf("NODE-%d", i),
			Alive: true,
			Data:  make(map[string]string),
		}

		nodes = append(nodes, node)
	}
}

//////////////////////////////////////////////////////
// WRITE TO NODE
//////////////////////////////////////////////////////

func write(
	node *Node,
	key string,
	value string,
) {

	if !node.Alive {

		log.Printf(
			"%s DOWN WRITE SKIPPED",
			node.ID,
		)

		return
	}

	node.Mutex.Lock()

	node.Data[key] = value

	node.Mutex.Unlock()

	log.Printf(
		"%s STORED %s=%s",
		node.ID,
		key,
		value,
	)
}

//////////////////////////////////////////////////////
// ASYNC REPLICATION
//////////////////////////////////////////////////////

func replicateAsync(
	source *Node,
	key string,
	value string,
) {

	for _, node := range nodes {

		if node.ID == source.ID {
			continue
		}

		go func(target *Node) {

			//////////////////////////////////////////////////
			// NETWORK LATENCY
			//////////////////////////////////////////////////

			time.Sleep(2 * time.Second)

			write(
				target,
				key,
				value,
			)

		}(node)
	}
}

//////////////////////////////////////////////////////
// CLIENT WRITE
//////////////////////////////////////////////////////

func clientWrite(
	key string,
	value string,
) {

	//////////////////////////////////////////////////
	// ANY NODE ACCEPTS WRITE
	//////////////////////////////////////////////////

	coordinator := nodes[0]

	log.Printf(
		"CLIENT WRITE RECEIVED BY %s",
		coordinator.ID,
	)

	//////////////////////////////////////////////////
	// LOCAL WRITE
	//////////////////////////////////////////////////

	write(
		coordinator,
		key,
		value,
	)

	//////////////////////////////////////////////////
	// ASYNC REPLICATION
	//////////////////////////////////////////////////

	go replicateAsync(
		coordinator,
		key,
		value,
	)
}

//////////////////////////////////////////////////////
// NETWORK PARTITION
//////////////////////////////////////////////////////

func networkPartition() {

	time.Sleep(5 * time.Second)

	log.Println(
		"NETWORK PARTITION NODE-2 DISCONNECTED",
	)

	nodes[1].Alive = false

	time.Sleep(10 * time.Second)

	log.Println(
		"NODE-2 RECONNECTED",
	)

	nodes[1].Alive = true
}

//////////////////////////////////////////////////////
// REPAIR MECHANISM
//////////////////////////////////////////////////////

func repairNode() {

	for {

		time.Sleep(5 * time.Second)

		if !nodes[1].Alive {
			continue
		}

		//////////////////////////////////////////////////
		// REPAIR NODE-2 FROM NODE-1
		//////////////////////////////////////////////////

		source := nodes[0]
		target := nodes[1]

		source.Mutex.Lock()

		for key, value := range source.Data {

			target.Data[key] = value
		}

		source.Mutex.Unlock()

		log.Println(
			"READ REPAIR COMPLETED FOR NODE-2",
		)
	}
}

//////////////////////////////////////////////////////
// CLUSTER STATUS
//////////////////////////////////////////////////////

func clusterStatus() {

	for {

		time.Sleep(3 * time.Second)

		fmt.Println()
		fmt.Println("================================")

		for _, node := range nodes {

			fmt.Printf(
				"%s ALIVE=%v DATA=%v\n",
				node.ID,
				node.Alive,
				node.Data,
			)
		}

		fmt.Println("================================")
		fmt.Println()
	}
}

//////////////////////////////////////////////////////
// MAIN
//////////////////////////////////////////////////////

func main() {

	createCluster()

	go networkPartition()

	go repairNode()

	go clusterStatus()

	//////////////////////////////////////////////////
	// CLIENT WRITES
	//////////////////////////////////////////////////

	clientWrite(
		"user1",
		"balance-5000",
	)

	time.Sleep(4 * time.Second)

	clientWrite(
		"user2",
		"balance-9000",
	)

	time.Sleep(4 * time.Second)

	clientWrite(
		"user3",
		"balance-12000",
	)

	select {}
}
