

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

		resp, err := client.Get("http://localhost:9201/heartbeat")

		if err != nil {

			log.Println("Leader NODE-1 DOWN")

			log.Println("NODE-3 Waiting For New Election")

			break
		}

		resp.Body.Close()

		log.Println("NODE-3 Received Heartbeat")

		time.Sleep(3 * time.Second)
	}
}

func main() {

	go checkLeader()

	log.Println("NODE-3 Running On Port 9203")

	http.ListenAndServe(":9203", nil)
}