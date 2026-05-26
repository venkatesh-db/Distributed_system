package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// curl http://localhost:8080/pay

//////////////////////////////////////////////////////
// SIDECAR PROXY
//////////////////////////////////////////////////////

type SidecarProxy struct {
	Name string
}

//////////////////////////////////////////////////////
// MTLS CLIENT
//////////////////////////////////////////////////////

func createMTLSClient() *http.Client {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}

	return client
}

//////////////////////////////////////////////////////
// TRAFFIC ROUTING
//////////////////////////////////////////////////////

func (s *SidecarProxy) routeTraffic() string {

	//////////////////////////////////////////////////
	// LOAD BALANCING
	//////////////////////////////////////////////////

	services := []string{
		"http://localhost:9001/process",
		"http://localhost:9002/process",
	}

	index := rand.Intn(len(services))

	return services[index]
}

//////////////////////////////////////////////////////
// RETRY LOGIC
//////////////////////////////////////////////////////

func (s *SidecarProxy) callService() {

	target := s.routeTraffic()

	log.Printf(
		"[SIDECAR] ROUTING REQUEST TO %s",
		target,
	)

	client := createMTLSClient()

	//////////////////////////////////////////////////
	// RETRIES
	//////////////////////////////////////////////////

	for retry := 1; retry <= 3; retry++ {

		resp, err := client.Get(target)

		if err != nil {

			log.Printf(
				"[SIDECAR] RETRY=%d ERROR=%v",
				retry,
				err,
			)

			time.Sleep(1 * time.Second)

			continue
		}

		body, _ := io.ReadAll(resp.Body)

		resp.Body.Close()

		log.Printf(
			"[SIDECAR] RESPONSE=%s",
			string(body),
		)

		return
	}

	log.Println(
		"[SIDECAR] SERVICE FAILED AFTER RETRIES",
	)
}

//////////////////////////////////////////////////////
// PAYMENT SERVICE
//////////////////////////////////////////////////////

func paymentService() {

	proxy := SidecarProxy{
		Name: "payment-sidecar",
	}

	http.HandleFunc("/pay", func(w http.ResponseWriter, r *http.Request) {

		log.Println(
			"[PAYMENT] PAYMENT REQUEST RECEIVED",
		)

		//////////////////////////////////////////////////
		// SIDECAR HANDLES NETWORKING
		//////////////////////////////////////////////////

		proxy.callService()

		fmt.Fprintf(
			w,
			"PAYMENT SUCCESS",
		)
	})

	log.Println(
		"PAYMENT SERVICE RUNNING ON 8080",
	)

	http.ListenAndServe(":8080", nil)
}

//////////////////////////////////////////////////////
// NOTIFICATION SERVICE - INSTANCE 1
//////////////////////////////////////////////////////

func notificationService1() {

	mux := http.NewServeMux()

	mux.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {

		log.Println(
			"[NOTIFICATION-1] REQUEST RECEIVED",
		)

		time.Sleep(1 * time.Second)

		fmt.Fprintf(
			w,
			"NOTIFICATION-1 SUCCESS",
		)
	})

	log.Println(
		"NOTIFICATION-1 RUNNING ON 9001",
	)

	http.ListenAndServe(":9001", mux)
}

//////////////////////////////////////////////////////
// NOTIFICATION SERVICE - INSTANCE 2
//////////////////////////////////////////////////////

func notificationService2() {

	mux := http.NewServeMux()

	mux.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {

		log.Println(
			"[NOTIFICATION-2] REQUEST RECEIVED",
		)

		//////////////////////////////////////////////////
		// RANDOM FAILURE
		//////////////////////////////////////////////////

		if rand.Intn(100) < 40 {

			log.Println(
				"[NOTIFICATION-2] FAILURE",
			)

			time.Sleep(4 * time.Second)

			http.Error(
				w,
				"SERVICE FAILURE",
				500,
			)

			return
		}

		time.Sleep(2 * time.Second)

		fmt.Fprintf(
			w,
			"NOTIFICATION-2 SUCCESS",
		)
	})

	log.Println(
		"NOTIFICATION-2 RUNNING ON 9002",
	)

	http.ListenAndServe(":9002", mux)
}

//////////////////////////////////////////////////////
// MAIN
//////////////////////////////////////////////////////

func main() {

	rand.Seed(time.Now().UnixNano())

	go notificationService1()

	go notificationService2()

	go paymentService()

	select {}
}
