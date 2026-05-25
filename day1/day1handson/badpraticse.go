package main

import (
	"fmt"
	"math/rand"
	"time"
)

func paymentService(
	requestChannel chan int,
	responseChannel chan string,
) {

	for {

		request := <-requestChannel

		fmt.Println(
			"PAYMENT SERVICE PROCESSING:",
			request,
		)

		time.Sleep(
			time.Duration(rand.Intn(5)+1) * time.Second,
		)

		responseChannel <- fmt.Sprintf(
			"PAYMENT SUCCESS %d",
			request,
		)
	}
}

func inventoryService(
	requestChannel chan int,
	responseChannel chan string,
) {

	for {

		request := <-requestChannel

		fmt.Println(
			"INVENTORY SERVICE PROCESSING:",
			request,
		)

		time.Sleep(
			time.Duration(rand.Intn(5)+1) * time.Second,
		)

		responseChannel <- fmt.Sprintf(
			"INVENTORY UPDATED %d",
			request,
		)
	}
}

func notificationService(
	responseChannel chan string,
) {

	for {

		message := <-responseChannel

		fmt.Println(
			"NOTIFICATION SENT:",
			message,
		)
	}
}

func main() {

	requestChannel := make(chan int)

	responseChannel := make(chan string)

	// TOO MANY UNCONTROLLED GOROUTINES

	for i := 1; i <= 10; i++ {

		go paymentService(
			requestChannel,
			responseChannel,
		)

		go inventoryService(
			requestChannel,
			responseChannel,
		)
	}

	go notificationService(responseChannel)

	// PRODUCER

	for requestID := 1; requestID <= 100; requestID++ {

		requestChannel <- requestID
	}

	select {}
}
