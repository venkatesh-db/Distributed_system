

package main

import (
	"io"
	"log"
	"net/http"
)

func gatewayHandler(w http.ResponseWriter, r *http.Request) {

	resp, err := http.Get(
		"http://localhost:9502/create-order",
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	w.Write(body)
}

func main() {

	http.HandleFunc("/order", gatewayHandler)

	log.Println("Gateway Running On Port 9500")

	http.ListenAndServe(":9500", nil)
}

