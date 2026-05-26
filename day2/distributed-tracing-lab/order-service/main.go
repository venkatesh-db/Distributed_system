package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initTracer() *sdktrace.TracerProvider {

	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(
				"http://jaeger:14268/api/traces",
			),
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
				semconv.ServiceName("order-service"),
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

	r.Use(otelgin.Middleware("order-service"))

	r.GET("/order", func(c *gin.Context) {

		time.Sleep(2 * time.Second)

		c.JSON(200, gin.H{
			"message": "Order Created Successfully",
		})
	})

	log.Println(
		"Order Service Running On Port 8082",
	)

	r.Run(":8082")
}
