package provider

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/b-j-roberts/foc-engine/internal/config"
	"github.com/gorilla/websocket"
)

func ConnectStarknetWebSocket(processStarknetEventData func([]byte)) (*websocket.Conn, error) {
	// Connect to the WebSocket server
	wsURL := "ws://" + config.Conf.Rpc.Host + "/ws"
	u, err := url.Parse(wsURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket:", err)
		StarknetProvider = &Provider{
			RpcHost:       config.Conf.Rpc.Host,
			WebSocketConn: nil,
		}
		return nil, err
	}

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Error reading message from WebSocket:", err)
				return
			}
			ProcessWebSocketMessage(message, processStarknetEventData) // TODO: Refactor func param
		}
	}()
	fmt.Println("Connected to WebSocket server at", u.String())

	return conn, nil
}

// TODO: Can we include more here?
type StarknetWsResponse struct {
	ID      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
}

func ProcessWebSocketMessage(message []byte, processStarknetEventData func([]byte)) {
	var response StarknetWsResponse
	err := json.Unmarshal(message, &response)
	if err != nil {
		fmt.Println("Error unmarshalling WebSocket message:", err)
		return
	}
	switch response.Method {
	case "":
		fmt.Println("Received empty msg:", string(message))
	case "starknet_subscribeNewHeads":
		// TODO
		fmt.Println("Received new head subscription message:", string(message))
	case "starknet_subscriptionEvents":
		fmt.Println("Received event subscription message:", string(message))
		processStarknetEventData(message)
	default:
		fmt.Println("Unknown WebSocket message method:", response.Method)
	}
}

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
}

func SubscribeEvents(address string) error {
	// TODO: Block number argument
	call := StarknetRpcCall{
		ID:      1,
		Jsonrpc: "2.0",
		Method:  "starknet_subscribeEvents",
		Params: map[string]interface{}{
			"block_id": map[string]interface{}{
				"block_number": 0,
			},
			"from_address": address,
		},
	}
	// Convert the call to JSON
	callBytes, err := json.Marshal(call)
	if err != nil {
		fmt.Println("Error marshalling call to JSON:", err)
		return err
	}

	if StarknetProvider.WebSocketConn == nil {
		fmt.Println("WebSocket connection is nil")
		return fmt.Errorf("WebSocket connection is nil")
	}
	err = StarknetProvider.WebSocketConn.WriteMessage(websocket.TextMessage, callBytes)
	if err != nil {
		fmt.Println("Error writing message to WebSocket:", err)
		return err
	}
	fmt.Println("Message sent to WebSocket:", call)

	return nil
}
