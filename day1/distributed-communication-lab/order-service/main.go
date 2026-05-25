package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func createOrder(w http.ResponseWriter, r *http.Request) {

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get("http://localhost:9501/pay")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orderID := "ORD-1001"

	go func() {

		_, err := http.Get(
			"http://localhost:9503/notify?orderId=" + orderID,
		)

		if err != nil {
			log.Println(err)
		}

	}()

	fmt.Fprintf(
		w,
		"Order Created\nPayment Response: %s",
		string(body),
	)
}

func main() {

	http.HandleFunc("/create-order", createOrder)

	log.Println("Order Service Running On Port 9502")

	err := http.ListenAndServe(":9502", nil)

	if err != nil {
		log.Fatal(err)
	}
}
