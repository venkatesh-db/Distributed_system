

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type PaymentResponse struct {
	Message string `json:"message"`
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {

	time.Sleep(2 * time.Second)

	response := PaymentResponse{
		Message: "Payment Successful",
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

func main() {

	http.HandleFunc("/pay", paymentHandler)

	log.Println("Payment Service Running On Port 9501")

	http.ListenAndServe(":9501", nil)
}