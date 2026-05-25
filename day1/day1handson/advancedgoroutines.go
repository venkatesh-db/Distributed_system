package main

import (
	"fmt"
	"sync"
)

var global int = 45

func smiles(ch chan int, smilescount int) {

	fmt.Println("before cresting slcie")
	var slice []int = make([]int, 5)
	slice = append(slice, smilescount)
	fmt.Println(slice, <-ch)

}

func main() {

	channel := make(chan int)
	go smiles(channel, 10)
	var wg sync.WaitGroup

	fmt.Println("before writing in to channel")
	channel <- 10
	fmt.Print("end of main")

	wg.Add(1)
	go func() {

		fmt.Println("unamed goroutine ")
		wg.Done()
	}()

	wg.Wait()

}
