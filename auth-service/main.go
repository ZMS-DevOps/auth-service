package main

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	startup "github.com/mmmajder/zms-devops-auth-service/startup"
	cfg "github.com/mmmajder/zms-devops-auth-service/startup/config"
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stdin)
	log.SetOutput(os.Stderr)
	log.SetOutput(os.Stdout)
	config := cfg.NewConfig()

	producer, _ := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootstrapServers,
		"security.protocol": "sasl_plaintext",
		"sasl.mechanism":    "PLAIN",
		"sasl.username":     "user1",
		"sasl.password":     config.KafkaAuthPassword,
	})
	defer producer.Close()

	server := startup.NewServer(config)
	server.Start(producer)
}
