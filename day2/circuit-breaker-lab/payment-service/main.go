package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initTracer() *sdktrace.TracerProvider {

	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
		),
	)

	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName("payment-service"),
			),
		),
	)

	otel.SetTracerProvider(tp)

	return tp
}

func main() {

	tp := initTracer()

	defer tp.Shutdown(context.Background())

	hystrix.ConfigureCommand("notification-service", hystrix.CommandConfig{
		Timeout:               3000,
		MaxConcurrentRequests: 10,
		ErrorPercentThreshold: 25,
	})

	r := gin.Default()

	r.Use(otelgin.Middleware("payment-service"))

	r.GET("/process-payment", func(c *gin.Context) {

		client := http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}

		err := hystrix.Do("notification-service", func() error {

			req, _ := http.NewRequest(
				"GET",
				"http://notification-service:8082/send-notification",
				nil,
			)

			resp, err := client.Do(req)

			if err != nil {
				return err
			}

			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			log.Println(string(body))

			return nil

		}, func(err error) error {

			log.Println("Circuit Breaker Triggered")

			return nil
		})

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "Payment Processed Successfully",
		})
	})

	log.Println("Payment Service Running On Port 8081")

	r.Run(":8081")
}
