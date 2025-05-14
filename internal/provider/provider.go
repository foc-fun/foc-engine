package provider

import (
	"fmt"
	"net/url"

	"github.com/b-j-roberts/foc-engine/internal/config"
	"github.com/gorilla/websocket"
)

type Provider struct {
  RpcHost string
  WebSocketConn *websocket.Conn
}

var StarknetProvider *Provider

func InitProvider() error {
  // Connect to the WebSocket server
  wsURL := "ws://" + config.Conf.Rpc.Host + "/ws"
  u, err := url.Parse(wsURL)
  fmt.Println("Connecting to WebSocket server at", u.String())
  conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
  if err != nil {
    fmt.Println("Error connecting to WebSocket:", err)
    StarknetProvider = &Provider{
      RpcHost: config.Conf.Rpc.Host,
      WebSocketConn: nil,
    }
    return err
  }

  // Create a new Provider instance
  StarknetProvider = &Provider{
    RpcHost: config.Conf.Rpc.Host,
    WebSocketConn: conn,
  }
  fmt.Println("Connected to WebSocket server at", u.String())
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
