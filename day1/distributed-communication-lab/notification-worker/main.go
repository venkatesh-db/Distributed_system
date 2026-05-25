package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func notificationHandler(w http.ResponseWriter, r *http.Request) {

	orderID := r.URL.Query().Get("orderId")

	fmt.Println("Received Notification Request:", orderID)

	time.Sleep(5 * time.Second)

	fmt.Println("Notification Sent For Order:", orderID)

	fmt.Fprintf(w, "Notification Processed")
}

func main() {

	http.HandleFunc("/notify", notificationHandler)

	log.Println("Notification Worker Running On Port 9503")

	err := http.ListenAndServe(":9503", nil)

	if err != nil {
		log.Fatal(err)
	}
}
