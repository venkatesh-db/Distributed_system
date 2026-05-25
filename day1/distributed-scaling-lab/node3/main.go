
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintf(w, "Response From NODE-3")
	})

	log.Println("NODE-3 Running On Port 9103")

	err := http.ListenAndServe(":9103", nil)

	if err != nil {
		log.Fatal(err)
	}
}