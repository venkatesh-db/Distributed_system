

package main

import (
	"log"
	"net/http"
	"time"
)

var leaderAlive = true

func checkLeader() {

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	for {

		resp, err := client.Get("http://localhost:9201/heartbeat")

		if err != nil {

			log.Println("Leader NODE-1 DOWN")

			leaderAlive = false

			log.Println("NODE-2 Becoming NEW LEADER")

			break
		}

		resp.Body.Close()

		log.Println("NODE-2 Received Heartbeat")

		time.Sleep(3 * time.Second)
	}
}

func main() {

	go checkLeader()

	log.Println("NODE-2 Running On Port 9202")

	http.ListenAndServe(":9202", nil)
}