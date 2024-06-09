package main

import (
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

	server := startup.NewServer(config)
	server.Start()
}
