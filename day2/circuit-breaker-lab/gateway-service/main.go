package main

import (
	"context"
	"io"
	"log"
	"net/http"

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
				semconv.ServiceName("gateway-service"),
			),
		),
	)

	otel.SetTracerProvider(tp)

	return tp
}

func main() {

	tp := initTracer()

	defer tp.Shutdown(context.Background())

	r := gin.Default()

	r.Use(otelgin.Middleware("gateway-service"))

	r.GET("/pay", func(c *gin.Context) {

		client := http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}

		req, _ := http.NewRequest(
			"GET",
			"http://payment-service:8081/process-payment",
			nil,
		)

		resp, err := client.Do(req)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		c.String(200, string(body))
	})

	log.Println("Gateway Service Running On Port 8080")

	r.Run(":8080")
}
