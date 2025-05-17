package provider

import (
	"fmt"

	"github.com/b-j-roberts/foc-engine/internal/config"
	"github.com/gorilla/websocket"
)

type Provider struct {
  RpcHost string
  WebSocketConn *websocket.Conn
}

var StarknetProvider *Provider

func InitProvider(processStarknetEventData func([]byte)) error {
  conn, err := ConnectStarknetWebSocket(processStarknetEventData)
  if err != nil {
    fmt.Println("Error connecting to WebSocket:", err)
    return err
  }
  // Create a new Provider instance
  StarknetProvider = &Provider{
    RpcHost: config.Conf.Rpc.Host,
    WebSocketConn: conn,
  }

  return nil
}

func Close() {
  if StarknetProvider != nil && StarknetProvider.WebSocketConn != nil {
    err := StarknetProvider.WebSocketConn.Close()
    if err != nil {
      fmt.Println("Error closing WebSocket connection:", err)
    } else {
      fmt.Println("WebSocket connection closed")
    }
  }
}
