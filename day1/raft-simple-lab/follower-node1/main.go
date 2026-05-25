package main

import (
	"log"
	"net/http"
	"time"
)

func checkLeader() {

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	for {

		resp, err := client.Get(
			"http://localhost:9601/heartbeat",
		)

		if err != nil {

			log.Println("Leader DOWN")

			log.Println("FOLLOWER-1 Starting Election")

			log.Println("FOLLOWER-1 Became NEW LEADER")

			break
		}

		resp.Body.Close()

		log.Println("FOLLOWER-1 Received Heartbeat")

		time.Sleep(3 * time.Second)
	}
}

func main() {

	go checkLeader()

	log.Println("Follower-1 Running On Port 9602")

	err := http.ListenAndServe(":9602", nil)

	if err != nil {
		log.Fatal(err)
	}
}
