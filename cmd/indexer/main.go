package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/b-j-roberts/foc-engine/internal/config"
	"github.com/b-j-roberts/foc-engine/internal/db/mongo"
	"github.com/b-j-roberts/foc-engine/internal/provider"
	"github.com/b-j-roberts/foc-engine/internal/registry"
	"github.com/b-j-roberts/foc-engine/routes"
)

func main() {
	config.InitConfig()

	err := provider.InitProvider(registry.ProcessStarknetEventData, true)
	if err != nil {
		fmt.Println("Error initializing provider:", err)
		os.Exit(1)
	}
	defer provider.Close()

	if mongo.ShouldConnectMongo() {
		mongo.InitMongoDB()
	}

	routes.StartServer(config.Conf.Indexer.Host, config.Conf.Indexer.Port)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})
	for {
		select {
		case <-done:
			fmt.Println("Connection closed")
			return
		case <-interrupt:
			fmt.Println("Interrupt signal received, shutting down...")
			close(done)
			return
		default:
			// Do nothing, just keep the connection alive
		}
	}
}
