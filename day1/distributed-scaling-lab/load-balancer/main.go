package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var servers = []string{
	"http://localhost:9101/data",
	"http://localhost:9102/data",
	"http://localhost:9103/data",
}

var counter uint64

func main() {

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	http.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {

		index := atomic.AddUint64(&counter, 1)

		server := servers[index%uint64(len(servers))]

		resp, err := client.Get(server)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(
			w,
			"Request Routed To %s\nResponse: %s",
			server,
			string(body),
		)
	})

	log.Println("Load Balancer Running On Port 8190")

	err := http.ListenAndServe(":8190", nil)

	if err != nil {
		log.Fatal(err)
	}
}
