package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func process(w http.ResponseWriter, r *http.Request) {

	orderID := r.URL.Query().Get("orderId")

	log.Println("EMAIL Processing:", orderID)

	time.Sleep(8 * time.Second)

	log.Println("EMAIL Sent:", orderID)

	fmt.Fprintf(w, "EMAIL Completed")
}

func main() {

	http.HandleFunc("/process", process)

	log.Println("Email Consumer Running On Port 9711")

	http.ListenAndServe(":9711", nil)
}
