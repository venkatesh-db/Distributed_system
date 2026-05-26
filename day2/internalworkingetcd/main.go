package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

//////////////////////////////////////////////////////
// NODE STRUCTURE
//////////////////////////////////////////////////////

type Node struct {
	ID       int
	IsLeader bool
	Alive    bool
	Term     int
	LastBeat time.Time
	Data     map[string]string
	Mutex    sync.Mutex
}

//////////////////////////////////////////////////////
// CLUSTER
//////////////////////////////////////////////////////

var cluster []*Node

var clusterMutex sync.Mutex

//////////////////////////////////////////////////////
// CREATE CLUSTER
//////////////////////////////////////////////////////

func createCluster() {

	for i := 1; i <= 5; i++ {

		node := &Node{
			ID:       i,
			IsLeader: false,
			Alive:    true,
			Term:     1,
			Data:     make(map[string]string),
		}

		cluster = append(cluster, node)
	}

	cluster[0].IsLeader = true

	log.Printf(
		"INITIAL LEADER ELECTED NODE=%d",
		cluster[0].ID,
	)
}

//////////////////////////////////////////////////////
// GET CURRENT LEADER
//////////////////////////////////////////////////////

func getLeader() *Node {

	for _, node := range cluster {

		if node.IsLeader && node.Alive {
			return node
		}
	}

	return nil
}

//////////////////////////////////////////////////////
// HEARTBEAT SYSTEM
//////////////////////////////////////////////////////

func heartbeatSender() {

	for {

		leader := getLeader()

		if leader == nil {

			time.Sleep(2 * time.Second)

			continue
		}

		for _, node := range cluster {

			if node.ID == leader.ID {
				continue
			}

			if !node.Alive {
				continue
			}

			node.LastBeat = time.Now()

			log.Printf(
				"LEADER=%d SENT HEARTBEAT TO NODE=%d",
				leader.ID,
				node.ID,
			)
		}

		time.Sleep(2 * time.Second)
	}
}

//////////////////////////////////////////////////////
// LEADER FAILURE DETECTION
//////////////////////////////////////////////////////

func monitorLeader() {

	for {

		time.Sleep(3 * time.Second)

		leader := getLeader()

		if leader == nil {

			log.Println(
				"LEADER FAILURE DETECTED",
			)

			startElection()
		}
	}
}

//////////////////////////////////////////////////////
// LEADER ELECTION
//////////////////////////////////////////////////////

func startElection() {

	clusterMutex.Lock()

	defer clusterMutex.Unlock()

	var candidates []*Node

	for _, node := range cluster {

		if node.Alive {

			node.IsLeader = false

			candidates = append(
				candidates,
				node,
			)
		}
	}

	if len(candidates) == 0 {

		log.Println(
			"NO NODES AVAILABLE",
		)

		return
	}

	newLeaderIndex := rand.Intn(
		len(candidates),
	)

	newLeader := candidates[newLeaderIndex]

	newLeader.IsLeader = true

	newLeader.Term++

	log.Printf(
		"NEW LEADER ELECTED NODE=%d TERM=%d",
		newLeader.ID,
		newLeader.Term,
	)
}

//////////////////////////////////////////////////////
// QUORUM CHECK
//////////////////////////////////////////////////////

func quorumAvailable() bool {

	aliveNodes := 0

	for _, node := range cluster {

		if node.Alive {
			aliveNodes++
		}
	}

	//////////////////////////////////////////////////
	// MAJORITY REQUIRED
	//////////////////////////////////////////////////

	required := (len(cluster) / 2) + 1

	return aliveNodes >= required
}

//////////////////////////////////////////////////////
// STRONG CONSISTENCY WRITE
//////////////////////////////////////////////////////

func writeData(
	key string,
	value string,
) {

	leader := getLeader()

	if leader == nil {

		log.Println(
			"NO LEADER AVAILABLE",
		)

		return
	}

	//////////////////////////////////////////////////
	// CP SYSTEM REQUIRES QUORUM
	//////////////////////////////////////////////////

	if !quorumAvailable() {

		log.Println(
			"QUORUM LOST WRITE REJECTED",
		)

		return
	}

	log.Printf(
		"LEADER=%d PROCESSING WRITE",
		leader.ID,
	)

	ackCount := 1

	//////////////////////////////////////////////////
	// REPLICATE TO FOLLOWERS
	//////////////////////////////////////////////////

	for _, node := range cluster {

		if node.ID == leader.ID {
			continue
		}

		if !node.Alive {
			continue
		}

		time.Sleep(500 * time.Millisecond)

		node.Mutex.Lock()

		node.Data[key] = value

		node.Mutex.Unlock()

		ackCount++

		log.Printf(
			"NODE=%d REPLICATED DATA",
			node.ID,
		)
	}

	requiredAcks := (len(cluster) / 2) + 1

	if ackCount >= requiredAcks {

		leader.Mutex.Lock()

		leader.Data[key] = value

		leader.Mutex.Unlock()

		log.Printf(
			"WRITE COMMITTED KEY=%s VALUE=%s",
			key,
			value,
		)

	} else {

		log.Println(
			"WRITE FAILED QUORUM NOT MET",
		)
	}
}

//////////////////////////////////////////////////////
// NETWORK PARTITION SIMULATION
//////////////////////////////////////////////////////

func networkPartitionSimulation() {

	for {

		time.Sleep(15 * time.Second)

		leader := getLeader()

		if leader == nil {
			continue
		}

		log.Printf(
			"NETWORK PARTITION LEADER=%d DISCONNECTED",
			leader.ID,
		)

		leader.Alive = false
		leader.IsLeader = false

		time.Sleep(10 * time.Second)

		leader.Alive = true

		log.Printf(
			"NODE=%d REJOINED CLUSTER",
			leader.ID,
		)
	}
}

//////////////////////////////////////////////////////
// CLUSTER STATUS
//////////////////////////////////////////////////////

func clusterStatus() {

	for {

		time.Sleep(5 * time.Second)

		fmt.Println()
		fmt.Println("================================")

		for _, node := range cluster {

			fmt.Printf(
				"NODE=%d LEADER=%v ALIVE=%v TERM=%d\n",
				node.ID,
				node.IsLeader,
				node.Alive,
				node.Term,
			)
		}

		fmt.Println("================================")
		fmt.Println()
	}
}

//////////////////////////////////////////////////////
// CLIENT WRITES
//////////////////////////////////////////////////////

func clientWrites() {

	counter := 1

	for {

		key := fmt.Sprintf(
			"config-%d",
			counter,
		)

		value := fmt.Sprintf(
			"value-%d",
			counter,
		)

		writeData(
			key,
			value,
		)

		counter++

		time.Sleep(4 * time.Second)
	}
}

//////////////////////////////////////////////////////
// MAIN
//////////////////////////////////////////////////////

func main() {

	rand.Seed(time.Now().UnixNano())

	createCluster()

	go heartbeatSender()

	go monitorLeader()

	go networkPartitionSimulation()

	go clusterStatus()

	go clientWrites()

	select {}
}
