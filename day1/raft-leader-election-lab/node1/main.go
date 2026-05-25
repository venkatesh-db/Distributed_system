

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var isLeader = true

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {

	if isLeader {
		fmt.Fprintf(w, "leader-alive")
		return
	}

	http.Error(w, "not leader", 500)
}

func main() {

	http.HandleFunc("/heartbeat", heartbeatHandler)

	go func() {

		for {

			if isLeader {
				log.Println("NODE-1 Sending Heartbeat")
			}

			time.Sleep(2 * time.Second)
		}
	}()

	log.Println("NODE-1 Running On Port 9201")

	http.ListenAndServe(":9201", nil)
}