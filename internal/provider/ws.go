package provider

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func SubscribeNewHeads() {
  call := StarknetRpcCall{
    ID:      1,
    Jsonrpc: "2.0",
    Method:  "starknet_subscribeNewHeads",
    Params:  map[string]interface{}{},
  }
  // Convert the call to JSON
  callBytes, err := json.Marshal(call)
  if err != nil {
    fmt.Println("Error marshalling call to JSON:", err)
    return
  }

  err = StarknetProvider.WebSocketConn.WriteMessage(websocket.TextMessage, callBytes)
  if err != nil {
    fmt.Println("Error writing message to WebSocket:", err)
    return
  }
  fmt.Println("Message sent to WebSocket:", call)
  go func() {
    for {
      _, message, err := StarknetProvider.WebSocketConn.ReadMessage()
      if err != nil {
        fmt.Println("Error reading message from WebSocket:", err)
        return
      }
      fmt.Println("Received message from WebSocket:", string(message))
    }
  }()
  fmt.Println("WebSocket connection established for new heads subscription")
}

func SubscribeEvents() {
  call := StarknetRpcCall{
    ID:      1,
    Jsonrpc: "2.0",
    Method:  "starknet_subscribeEvents",
    Params:  map[string]interface{}{
      "block_id": map[string]interface{}{
        "block_number": 0,
      },
    },
  }
  // Convert the call to JSON
  callBytes, err := json.Marshal(call)
  if err != nil {
    fmt.Println("Error marshalling call to JSON:", err)
    return
  }
  
  err = StarknetProvider.WebSocketConn.WriteMessage(websocket.TextMessage, callBytes)
  if err != nil {
    fmt.Println("Error writing message to WebSocket:", err)
    return
  }
  fmt.Println("Message sent to WebSocket:", call)

  go func() {
    for {
      _, message, err := StarknetProvider.WebSocketConn.ReadMessage()
      if err != nil {
        fmt.Println("Error reading message from WebSocket:", err)
        return
      }
      fmt.Println("Received message from WebSocket:", string(message))
    }
  }()
  fmt.Println("WebSocket connection established for event subscription")
}
