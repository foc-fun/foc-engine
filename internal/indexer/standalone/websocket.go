package standalone

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketEventData represents the event data received via WebSocket
type WebSocketEventData struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		SubscriptionId string `json:"subscription_id"`
		Result         struct {
			BlockHash       string   `json:"block_hash"`
			BlockNumber     uint64   `json:"block_number"`
			FromAddress     string   `json:"from_address"`
			TransactionHash string   `json:"transaction_hash"`
			Keys            []string `json:"keys"`
			Data            []string `json:"data"`
		} `json:"result"`
	} `json:"params"`
}

// WebSocketResponse represents a generic WebSocket response
type WebSocketResponse struct {
	ID      int         `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
}

// connectWebSocket establishes a WebSocket connection to the Starknet node
func (idx *Indexer) connectWebSocket() (*websocket.Conn, error) {
	// Parse the RPC URL and convert to WebSocket URL
	rpcURL := idx.config.RPC
	
	// Replace https:// with wss:// or http:// with ws://
	var wsURL string
	if strings.HasPrefix(rpcURL, "https://") {
		wsURL = strings.Replace(rpcURL, "https://", "wss://", 1)
	} else if strings.HasPrefix(rpcURL, "http://") {
		wsURL = strings.Replace(rpcURL, "http://", "ws://", 1)
	} else {
		return nil, fmt.Errorf("invalid RPC URL format: %s", rpcURL)
	}
	
	// Parse URL
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WebSocket URL: %v", err)
	}
	
	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %v", err)
	}
	
	fmt.Printf("Connected to WebSocket at %s\n", u.String())
	return conn, nil
}

// subscribeToEvents subscribes to events for the configured contract
func (idx *Indexer) subscribeToEvents(conn *websocket.Conn) error {
	// Normalize contract address
	contractAddress := idx.normalizeAddress(idx.config.Contract)
	
	// Prepare subscription request
	subscribeCall := StarknetRpcCall{
		ID:      1,
		Jsonrpc: "2.0",
		Method:  "starknet_subscribeEvents",
		Params: map[string]interface{}{
			"from_block": map[string]interface{}{
				"block_number": idx.config.StartBlock,
			},
			"from_address": contractAddress,
			"keys": [][]string{
				{idx.eventSelector}, // First key is the event selector
			},
		},
	}
	
	// Send subscription request
	if err := conn.WriteJSON(subscribeCall); err != nil {
		return fmt.Errorf("failed to send subscription request: %v", err)
	}
	
	fmt.Printf("Subscribed to events from contract %s\n", contractAddress)
	fmt.Printf("Filtering for event selector: %s\n", idx.eventSelector)
	
	return nil
}

// handleWebSocketMessages processes incoming WebSocket messages
func (idx *Indexer) handleWebSocketMessages(conn *websocket.Conn) {
	for idx.running {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if !idx.running {
				return // Normal shutdown
			}
			fmt.Printf("Error reading WebSocket message: %v\n", err)
			
			// Try to reconnect
			time.Sleep(5 * time.Second)
			newConn, err := idx.connectWebSocket()
			if err != nil {
				fmt.Printf("Failed to reconnect: %v\n", err)
				continue
			}
			
			// Re-subscribe to events
			if err := idx.subscribeToEvents(newConn); err != nil {
				fmt.Printf("Failed to re-subscribe: %v\n", err)
				newConn.Close()
				continue
			}
			
			// Update connection and continue
			conn.Close()
			conn = newConn
			continue
		}
		
		// Process the message
		idx.processWebSocketMessage(message)
	}
}

// processWebSocketMessage processes a single WebSocket message
func (idx *Indexer) processWebSocketMessage(message []byte) {
	// First, try to parse as a generic response to check the method
	var response WebSocketResponse
	if err := json.Unmarshal(message, &response); err != nil {
		fmt.Printf("Error unmarshalling WebSocket message: %v\n", err)
		return
	}
	
	// Check if this is a subscription confirmation
	if response.Method == "" && response.Result != nil {
		fmt.Printf("Subscription confirmed: %v\n", response.Result)
		return
	}
	
	// Check if this is an event notification
	if response.Method == "starknet_subscriptionEvents" {
		// Parse as event data
		var eventData WebSocketEventData
		if err := json.Unmarshal(message, &eventData); err != nil {
			fmt.Printf("Error unmarshalling event data: %v\n", err)
			return
		}
		
		idx.processEvent(eventData)
	}
}

// processEvent processes a single event received via WebSocket
func (idx *Indexer) processEvent(wsEvent WebSocketEventData) {
	// Extract order key based on the configured index
	allValues := append(wsEvent.Params.Result.Keys, wsEvent.Params.Result.Data...)
	var orderKey string
	if idx.config.OrderBy >= 0 && idx.config.OrderBy < len(allValues) {
		orderKey = allValues[idx.config.OrderBy]
	} else {
		orderKey = fmt.Sprintf("%020d", wsEvent.Params.Result.BlockNumber)
	}
	
	// Create EventData
	event := EventData{
		BlockNumber:     wsEvent.Params.Result.BlockNumber,
		TransactionHash: wsEvent.Params.Result.TransactionHash,
		FromAddress:     wsEvent.Params.Result.FromAddress,
		Keys:            wsEvent.Params.Result.Keys,
		Data:            wsEvent.Params.Result.Data,
		Timestamp:       time.Now().Unix(),
		OrderKey:        orderKey,
	}
	
	// Store the event
	idx.storeEvents([]EventData{event})
	
	// Update current block
	if wsEvent.Params.Result.BlockNumber > idx.currentBlock {
		idx.currentBlock = wsEvent.Params.Result.BlockNumber
	}
	
	fmt.Printf("Indexed event at block %d (tx: %s)\n", event.BlockNumber, event.TransactionHash[:10]+"...")
	fmt.Printf("  Total events stored: %d\n", idx.GetEventCount())
}

// tryWebSocket attempts to start WebSocket-based indexing
func (idx *Indexer) tryWebSocket() error {
	fmt.Printf("Starting WebSocket indexer from block %d\n", idx.config.StartBlock)
	fmt.Printf("Indexing events from contract: %s\n", idx.config.Contract)
	fmt.Printf("Event: %s (selector: %s)\n", idx.config.Event, idx.eventSelector)
	
	// Connect to WebSocket
	conn, err := idx.connectWebSocket()
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()
	
	// Subscribe to events
	if err := idx.subscribeToEvents(conn); err != nil {
		return fmt.Errorf("failed to subscribe to events: %v", err)
	}
	
	// Handle incoming messages
	idx.handleWebSocketMessages(conn)
	
	return nil
}