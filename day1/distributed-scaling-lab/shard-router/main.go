package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"time"
)

var nodes = []string{
	"http://localhost:9101/data",
	"http://localhost:9102/data",
	"http://localhost:9103/data",
}

func getNode(userID string) string {

	hash := fnv.New32a()

	_, err := hash.Write([]byte(userID))

	if err != nil {
		log.Println(err)
	}

	index := hash.Sum32() % uint32(len(nodes))

	return nodes[index]
}

func main() {

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	http.HandleFunc("/route", func(w http.ResponseWriter, r *http.Request) {

		userID := r.URL.Query().Get("user")

		if userID == "" {
			http.Error(w, "user query parameter required", 400)
			return
		}

		node := getNode(userID)

		resp, err := client.Get(node)

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
			"User %s Routed To %s\nResponse: %s",
			userID,
			node,
			string(body),
		)
	})

	log.Println("Shard Router Running On Port 8180")

	err := http.ListenAndServe(":8180", nil)

	if err != nil {
		log.Fatal(err)
	}
}
