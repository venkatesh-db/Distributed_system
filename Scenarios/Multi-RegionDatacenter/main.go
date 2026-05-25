package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Region struct {
	Name    string
	Alive   bool
	Latency time.Duration
}

var (
	regions = map[string]*Region{
		"mumbai": {
			Name:    "Mumbai",
			Alive:   true,
			Latency: 1 * time.Second,
		},
		"hyderabad": {
			Name:    "Hyderabad",
			Alive:   true,
			Latency: 2 * time.Second,
		},
	}

	mutex sync.Mutex
)

func processRequest(region *Region) string {

	time.Sleep(region.Latency)

	return fmt.Sprintf(
		"Response From %s Region",
		region.Name,
	)
}

func loadBalancer(w http.ResponseWriter, r *http.Request) {

	mutex.Lock()

	primary := regions["mumbai"]
	secondary := regions["hyderabad"]

	mutex.Unlock()

	if primary.Alive {

		log.Println(
			"ROUTING TRAFFIC TO PRIMARY REGION: Mumbai",
		)

		response := processRequest(primary)

		fmt.Fprintf(w, response)

		return
	}

	if secondary.Alive {

		log.Println(
			"PRIMARY REGION DOWN - FAILOVER TO HYDERABAD",
		)

		response := processRequest(secondary)

		fmt.Fprintf(w, response)

		return
	}

	http.Error(
		w,
		"ALL REGIONS DOWN",
		http.StatusServiceUnavailable,
	)
}

func simulateRegionFailure() {

	for {

		time.Sleep(15 * time.Second)

		mutex.Lock()

		regions["mumbai"].Alive = false

		mutex.Unlock()

		log.Println("MUMBAI REGION FAILED")

		time.Sleep(15 * time.Second)

		mutex.Lock()

		regions["mumbai"].Alive = true

		mutex.Unlock()

		log.Println("MUMBAI REGION RECOVERED")
	}
}

func trafficGenerator() {

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	for {

		time.Sleep(2 * time.Second)

		go func() {

			resp, err := client.Get(
				"http://localhost:9904/api",
			)

			if err != nil {

				log.Println(
					"CLIENT REQUEST FAILED",
				)

				return
			}

			log.Printf(
				"CLIENT RECEIVED STATUS=%d",
				resp.StatusCode,
			)

			resp.Body.Close()
		}()
	}
}

func latencySpikeSimulation() {

	for {

		time.Sleep(20 * time.Second)

		mutex.Lock()

		regions["hyderabad"].Latency =
			time.Duration(rand.Intn(5)+1) * time.Second

		log.Printf(
			"HYDERABAD LATENCY SPIKE=%v",
			regions["hyderabad"].Latency,
		)

		mutex.Unlock()
	}
}

func metricsPrinter() {

	for {

		time.Sleep(5 * time.Second)

		mutex.Lock()

		log.Println("---------------------------")

		for _, region := range regions {

			log.Printf(
				"REGION=%s ALIVE=%v LATENCY=%v",
				region.Name,
				region.Alive,
				region.Latency,
			)
		}

		log.Println("---------------------------")

		mutex.Unlock()
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/api", loadBalancer)

	go simulateRegionFailure()

	go trafficGenerator()

	go latencySpikeSimulation()

	go metricsPrinter()

	log.Println(
		"MULTI REGION FAILOVER LAB RUNNING ON PORT 9904",
	)

	err := http.ListenAndServe(":9904", nil)

	if err != nil {
		log.Fatal(err)
	}
}
