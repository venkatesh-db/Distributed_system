package main

import "fmt"

func serve1(ch1 chan int, ch2 chan string) {

	msg := <-ch1
	fmt.Println("serve1 received:", msg)
	ch2 <- "Hello from serve1" + fmt.Sprint(msg)
	fmt.Println("serve1 goroutine")

}

func server2(ch1 chan int, ch2 chan string) {

	fmt.Println("serve2 goroutine")
}

func main() {

	ch1 := make(chan int)
	ch2 := make(chan string)

	go serve1(ch1, ch2)
	go server2(ch1, ch2)

	fmt.Println("before writing first channel")
	ch1 <- 10
	fmt.Println("after writing first channel")

}
