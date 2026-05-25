
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintf(w, "Response From NODE-2")
	})

	log.Println("NODE-2 Running On Port 9102")

	err := http.ListenAndServe(":9102", nil)

	if err != nil {
		log.Fatal(err)
	}
}