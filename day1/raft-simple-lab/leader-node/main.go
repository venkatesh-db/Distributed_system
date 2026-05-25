package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "leader-alive")
}

func main() {

	http.HandleFunc("/heartbeat", heartbeatHandler)

	go func() {

		for {

			log.Println("LEADER Sending Heartbeat")

			time.Sleep(2 * time.Second)
		}
	}()

	log.Println("Leader Node Running On Port 9601")

	err := http.ListenAndServe(":9601", nil)

	if err != nil {
		log.Fatal(err)
	}
}
