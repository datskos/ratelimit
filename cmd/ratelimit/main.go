package main

import (
	"log"

	"github.com/datskos/ratelimit/pkg/config"
	"github.com/datskos/ratelimit/pkg/server"
)

func main() {
	config := config.NewConfig()
	server, err := server.NewServer(config)
	if err != nil {
		log.Fatalf("unable to create server: %s", err)
	}

	log.Printf("starting Server on port=%d", config.Port)
	err = server.Start()
	if err != nil {
		log.Fatal("error starting server: %s", err)
	}
}
