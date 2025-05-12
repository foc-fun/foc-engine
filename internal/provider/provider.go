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

func InitProvider() {
  // Connect to the WebSocket server
  u := url.URL{Scheme: "ws", Host: config.Conf.Rpc.Host, Path: "/ws"}
  conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
  if err != nil {
    fmt.Println("Error connecting to WebSocket:", err)
    return
  }

  // Create a new Provider instance
  StarknetProvider = &Provider{
    RpcHost: config.Conf.Rpc.Host,
    WebSocketConn: conn,
  }
  fmt.Println("Connected to WebSocket server at", u.String())
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
