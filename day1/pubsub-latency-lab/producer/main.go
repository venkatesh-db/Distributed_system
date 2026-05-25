package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func createOrder(w http.ResponseWriter, r *http.Request) {

	orderID := fmt.Sprintf(
		"ORD-%d",
		time.Now().UnixNano(),
	)

	log.Println("Publishing Order:", orderID)

	go func() {

		_, err := http.Get(
			"http://localhost:9711/process?orderId=" + orderID,
		)

		if err != nil {
			log.Println("EMAIL ERROR:", err)
		}
	}()

	go func() {

		_, err := http.Get(
			"http://localhost:9712/process?orderId=" + orderID,
		)

		if err != nil {
			log.Println("INVENTORY ERROR:", err)
		}
	}()

	go func() {

		_, err := http.Get(
			"http://localhost:9713/process?orderId=" + orderID,
		)

		if err != nil {
			log.Println("ANALYTICS ERROR:", err)
		}
	}()

	fmt.Fprintf(
		w,
		"Order Published Successfully: %s",
		orderID,
	)
}

func main() {

	http.HandleFunc("/create-order", createOrder)

	log.Println("Producer Running On Port 9700")

	err := http.ListenAndServe(":9700", nil)

	if err != nil {
		log.Fatal(err)
	}
}
