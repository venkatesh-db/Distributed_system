package main

import (
	"log"
	"math/rand"
	"time"
)

type Node struct {
	ID       string
	IsLeader bool
	Alive    bool
}

var cluster []Node

func electLeader() {

	leaderIndex := rand.Intn(len(cluster))

	for i := range cluster {
		cluster[i].IsLeader = false
	}

	cluster[leaderIndex].IsLeader = true

	log.Printf("NEW LEADER ELECTED=%s", cluster[leaderIndex].ID)
}

func simulateFailure() {

	for {

		time.Sleep(10 * time.Second)

		for i := range cluster {

			if cluster[i].IsLeader {

				log.Printf("LEADER FAILED=%s", cluster[i].ID)

				cluster[i].Alive = false
				cluster[i].IsLeader = false

				electLeader()

				break
			}
		}
	}
}

func clusterMonitor() {

	for {

		time.Sleep(3 * time.Second)

		for _, node := range cluster {

			log.Printf(
				"NODE=%s LEADER=%v ALIVE=%v",
				node.ID,
				node.IsLeader,
				node.Alive,
			)
		}
	}
}

func main() {

	cluster = []Node{
		{ID: "NODE-1", Alive: true},
		{ID: "NODE-2", Alive: true},
		{ID: "NODE-3", Alive: true},
	}

	rand.Seed(time.Now().UnixNano())

	electLeader()

	go simulateFailure()

	go clusterMonitor()

	select {}
}
