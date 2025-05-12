package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/b-j-roberts/foc-engine/internal/config"
	"github.com/b-j-roberts/foc-engine/internal/provider"
)

func main() {
  config.InitConfig()

  provider.InitProvider()
  defer provider.Close()
  provider.SubscribeEvents()

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
