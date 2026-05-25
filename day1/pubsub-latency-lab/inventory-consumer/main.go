package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func process(w http.ResponseWriter, r *http.Request) {

	orderID := r.URL.Query().Get("orderId")

	log.Println("INVENTORY Updating:", orderID)

	time.Sleep(2 * time.Second)

	log.Println("INVENTORY Updated:", orderID)

	fmt.Fprintf(w, "INVENTORY Completed")
}

func main() {

	http.HandleFunc("/process", process)

	log.Println("Inventory Consumer Running On Port 9712")

	http.ListenAndServe(":9712", nil)
}
