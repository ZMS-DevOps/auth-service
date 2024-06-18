package main

import (
	"context"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	startup "github.com/mmmajder/zms-devops-auth-service/startup"
	cfg "github.com/mmmajder/zms-devops-auth-service/startup/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"log"
	"os"
	"time"
)

func initJaegerTracer(jaegerHost string) (*sdktrace.TracerProvider, error) {
	log.Printf("Initializing tracing to jaeger at %s\n", jaegerHost)
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerHost)))
	if err != nil {
		return nil, err
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(domain.ServiceName),
		)),
	), nil
}

func initPromtailClient(lokiHost string) (promtail.Client, error) {
	labels := "{source=\"" + domain.ServiceName + "\",service_name=\"" + "\"}"
	conf := promtail.ClientConfig{
		PushURL:            lokiHost,
		Labels:             labels,
		BatchWait:          5 * time.Second,
		BatchEntriesNumber: 10000,
		SendLevel:          promtail.INFO,
		PrintLevel:         promtail.ERROR,
	}

	return promtail.NewClientJson(conf)
}

func main() {
	log.SetOutput(os.Stdin)
	log.SetOutput(os.Stderr)
	log.SetOutput(os.Stdout)
	config := cfg.NewConfig()

	var err error
	tp, err := initJaegerTracer(config.JaegerHost)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	loki, err := initPromtailClient(config.LokiHost)

	if err != nil {
		log.Printf("promtail.NewClient: %s\n", err)
	}

	producer, _ := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootstrapServers,
		"security.protocol": "sasl_plaintext",
		"sasl.mechanism":    "PLAIN",
		"sasl.username":     "user1",
		"sasl.password":     config.KafkaAuthPassword,
	})
	defer producer.Close()

	server := startup.NewServer(config, tp, loki)
	server.Start(producer)
	loki.Shutdown()
}
