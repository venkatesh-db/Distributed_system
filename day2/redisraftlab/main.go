package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

//////////////////////////////////////////////////////
// NODE
//////////////////////////////////////////////////////

type Node struct {
	ID            int
	IsLeader      bool
	Alive         bool
	CurrentTerm   int
	LastHeartbeat time.Time
	Mutex         sync.Mutex
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
			ID:            i,
			IsLeader:      false,
			Alive:         true,
			CurrentTerm:   1,
			LastHeartbeat: time.Now(),
		}

		cluster = append(
			cluster,
			node,
		)
	}

	//////////////////////////////////////////////////
	// INITIAL LEADER
	//////////////////////////////////////////////////

	cluster[0].IsLeader = true

	log.Printf(
		"INITIAL LEADER ELECTED NODE=%d",
		cluster[0].ID,
	)
}

//////////////////////////////////////////////////////
// GET LEADER
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
// HEARTBEATS
//////////////////////////////////////////////////////

func sendHeartbeats() {

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

			node.LastHeartbeat = time.Now()

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
// MONITOR LEADER
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
// RAFT LEADER ELECTION
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
			"NO AVAILABLE NODES",
		)

		return
	}

	//////////////////////////////////////////////////
	// RANDOM LEADER ELECTION
	//////////////////////////////////////////////////

	newLeaderIndex := rand.Intn(
		len(candidates),
	)

	newLeader := candidates[newLeaderIndex]

	newLeader.IsLeader = true

	newLeader.CurrentTerm++

	log.Printf(
		"NEW LEADER ELECTED NODE=%d TERM=%d",
		newLeader.ID,
		newLeader.CurrentTerm,
	)
}

//////////////////////////////////////////////////////
// LEADER CRASH SIMULATION
//////////////////////////////////////////////////////

func simulateLeaderCrash() {

	for {

		time.Sleep(15 * time.Second)

		leader := getLeader()

		if leader == nil {
			continue
		}

		log.Printf(
			"LEADER NODE=%d CRASHED",
			leader.ID,
		)

		leader.Alive = false
		leader.IsLeader = false

		//////////////////////////////////////////////////
		// RECOVERY
		//////////////////////////////////////////////////

		go recoverNode(leader)
	}
}

//////////////////////////////////////////////////////
// NODE RECOVERY
//////////////////////////////////////////////////////

func recoverNode(
	node *Node,
) {

	time.Sleep(10 * time.Second)

	node.Alive = true

	log.Printf(
		"NODE=%d REJOINED CLUSTER",
		node.ID,
	)
}

//////////////////////////////////////////////////////
// CLUSTER STATUS
//////////////////////////////////////////////////////

func printClusterStatus() {

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
				node.CurrentTerm,
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

		time.Sleep(4 * time.Second)

		leader := getLeader()

		if leader == nil {

			log.Println(
				"WRITE FAILED NO LEADER",
			)

			continue
		}

		log.Printf(
			"WRITE REQUEST PROCESSED BY LEADER=%d DATA=config-%d",
			leader.ID,
			counter,
		)

		counter++
	}
}

//////////////////////////////////////////////////////
// MAIN
//////////////////////////////////////////////////////

func main() {

	rand.Seed(time.Now().UnixNano())

	createCluster()

	go sendHeartbeats()

	go monitorLeader()

	go simulateLeaderCrash()

	go printClusterStatus()

	go clientWrites()

	select {}
}
