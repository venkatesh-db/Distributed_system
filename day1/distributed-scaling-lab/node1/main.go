
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintf(w, "Response From NODE-1")
	})

	log.Println("NODE-1 Running On Port 9101")

	err := http.ListenAndServe(":9101", nil)

	if err != nil {
		log.Fatal(err)
	}
}