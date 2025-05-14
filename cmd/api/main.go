package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/b-j-roberts/foc-engine/internal/config"
	"github.com/b-j-roberts/foc-engine/internal/provider"
)

func main() {
  config.InitConfig()

  // Sleep for 10 seconds
  time.Sleep(10 * time.Second)
  err := provider.InitProvider()
  if err != nil {
    fmt.Println("Error initializing provider:", err)
    os.Exit(1)
  }
  defer provider.Close()

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
