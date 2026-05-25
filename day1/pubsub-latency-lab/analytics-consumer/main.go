package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func process(w http.ResponseWriter, r *http.Request) {

	orderID := r.URL.Query().Get("orderId")

	log.Println("ANALYTICS Processing:", orderID)

	time.Sleep(1 * time.Second)

	log.Println("ANALYTICS Completed:", orderID)

	fmt.Fprintf(w, "ANALYTICS Completed")
}

func main() {

	http.HandleFunc("/process", process)

	log.Println("Analytics Consumer Running On Port 9713")

	err := http.ListenAndServe(":9713", nil)

	if err != nil {
		log.Fatal(err)
	}
}
