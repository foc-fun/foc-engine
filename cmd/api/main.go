package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/b-j-roberts/foc-engine/internal/config"
	"github.com/b-j-roberts/foc-engine/internal/db/mongo"
	"github.com/b-j-roberts/foc-engine/routes"
)

func main() {
	config.InitConfig()

	// TODO: Remove this after testing
	// Sleep for 10 seconds
	time.Sleep(10 * time.Second)

	if mongo.ShouldConnectMongo() {
		mongo.InitMongoDB()
	}

	routes.StartServer(config.Conf.Api.Host, config.Conf.Api.Port)

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
