package standalone

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/NethermindEth/starknet.go/utils"
)

// StarknetRpcCall represents a JSON-RPC call to Starknet
type StarknetRpcCall struct {
	ID      int         `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// StarknetRpcResponse represents a JSON-RPC response from Starknet
type StarknetRpcResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
}

// EventFilter represents the filter for querying events
type EventFilter struct {
	FromBlock   *BlockID     `json:"from_block,omitempty"`
	ToBlock     *BlockID     `json:"to_block,omitempty"`
	Address     string       `json:"address,omitempty"`
	Keys        [][]string   `json:"keys,omitempty"`
}

// BlockID represents a block identifier
type BlockID struct {
	BlockNumber *uint64 `json:"block_number,omitempty"`
	BlockHash   *string `json:"block_hash,omitempty"`
}

// Event represents a Starknet event
type Event struct {
	BlockHash       string   `json:"block_hash"`
	BlockNumber     uint64   `json:"block_number"`
	FromAddress     string   `json:"from_address"`
	TransactionHash string   `json:"transaction_hash"`
	Keys            []string `json:"keys"`
	Data            []string `json:"data"`
}

// EventsResult represents the result of get_events call
type EventsResult struct {
	Events []Event `json:"events"`
}

// getLatestBlockNumber gets the latest block number from the RPC
func (idx *Indexer) getLatestBlockNumber() (uint64, error) {
	call := StarknetRpcCall{
		ID:      1,
		Jsonrpc: "2.0",
		Method:  "starknet_blockNumber",
		Params:  []interface{}{},
	}
	
	response, err := idx.makeRPCCall(call)
	if err != nil {
		return 0, err
	}
	
	// Parse the block number from response
	// The result can be either a float64 or an int
	var blockNum uint64
	switch v := response.Result.(type) {
	case float64:
		blockNum = uint64(v)
	case int:
		blockNum = uint64(v)
	case int64:
		blockNum = uint64(v)
	default:
		return 0, fmt.Errorf("unexpected response type for block number: %T", response.Result)
	}
	
	return blockNum, nil
}

// getEventsInRange retrieves events in a block range with continuation token support
func (idx *Indexer) getEventsInRange(fromBlock, toBlock uint64, continuationToken string) ([]EventData, string, error) {
	// Normalize contract address
	contractAddress := idx.normalizeAddress(idx.config.Contract)
	
	// Use cached event selector
	eventSelector := idx.eventSelector
	
	// Build the parameters object according to Starknet JSON-RPC spec
	params := map[string]interface{}{
		"from_block": &BlockID{BlockNumber: &fromBlock},
		"to_block":   &BlockID{BlockNumber: &toBlock},
		"chunk_size": 1000,
	}
	
	// Add optional filter parameters
	if contractAddress != "" {
		params["address"] = contractAddress
	}
	if strings.HasPrefix(eventSelector, "0x") {
		params["keys"] = [][]string{{eventSelector}}
	}
	if continuationToken != "" {
		params["continuation_token"] = continuationToken
	}
	
	call := StarknetRpcCall{
		ID:      1,
		Jsonrpc: "2.0",
		Method:  "starknet_getEvents",
		Params:  []interface{}{params},
	}
	
	response, err := idx.makeRPCCall(call)
	if err != nil {
		return nil, "", err
	}
	
	// Parse the response result structure
	resultMap, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("unexpected response result format")
	}
	
	// Extract events array
	eventsInterface, ok := resultMap["events"]
	if !ok {
		return nil, "", fmt.Errorf("no events field in response")
	}
	
	eventsJson, err := json.Marshal(eventsInterface)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal events: %v", err)
	}
	
	var events []Event
	if err := json.Unmarshal(eventsJson, &events); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal events: %v", err)
	}
	
	// Extract continuation token (if present)
	var nextToken string
	if token, ok := resultMap["continuation_token"]; ok {
		if tokenStr, ok := token.(string); ok {
			nextToken = tokenStr
		}
	}
	
	// Convert to EventData
	var eventData []EventData
	for _, event := range events {
		orderKey := idx.extractOrderKey(event)
		uniqueKey := idx.extractUniqueKey(event)
		
		data := EventData{
			BlockNumber:     event.BlockNumber,
			TransactionHash: event.TransactionHash,
			FromAddress:     event.FromAddress,
			Keys:            padHexArray(event.Keys),
			Data:            padHexArray(event.Data),
			Timestamp:       time.Now().Unix(),
			OrderKey:        orderKey,
			UniqueKey:       uniqueKey,
		}
		eventData = append(eventData, data)
	}
	
	return eventData, nextToken, nil
}

// getEventsAtBlock retrieves events at a specific block (legacy function for backward compatibility)
func (idx *Indexer) getEventsAtBlock(blockNumber uint64) ([]EventData, error) {
	events, _, err := idx.getEventsInRange(blockNumber, blockNumber, "")
	return events, err
}

// makeRPCCall makes a JSON-RPC call to the Starknet node
func (idx *Indexer) makeRPCCall(call StarknetRpcCall) (*StarknetRpcResponse, error) {
	callBytes, err := json.Marshal(call)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC call: %v", err)
	}
	
	// Convert WebSocket URLs to HTTP for RPC calls
	rpcURL := idx.config.RPC
	if strings.HasPrefix(rpcURL, "wss://") {
		rpcURL = strings.Replace(rpcURL, "wss://", "https://", 1)
	} else if strings.HasPrefix(rpcURL, "ws://") {
		rpcURL = strings.Replace(rpcURL, "ws://", "http://", 1)
	}
	
	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(callBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to make RPC call: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	var response StarknetRpcResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %v", response.Error)
	}
	
	return &response, nil
}

// normalizeAddress normalizes a Starknet address to 0x-prefixed 64-char hex
func (idx *Indexer) normalizeAddress(address string) string {
	// Remove 0x prefix if present
	if strings.HasPrefix(address, "0x") {
		address = address[2:]
	}
	
	// Pad with leading zeros to 64 characters
	address = fmt.Sprintf("%064s", address)
	
	// Add 0x prefix
	return "0x" + address
}

// getEventSelector calculates the event selector from event name
func (idx *Indexer) getEventSelector(eventName string) string {
	// If it's already a hex selector, use it as-is
	if strings.HasPrefix(eventName, "0x") {
		return eventName
	}
	
	// Use starknet.go's GetSelectorFromName to compute the selector
	selector := utils.GetSelectorFromName(eventName)
	selectorHex := fmt.Sprintf("0x%064x", selector)
	
	return selectorHex
}

// extractOrderKey extracts the key value to order by from the event
func (idx *Indexer) extractOrderKey(event Event) string {
	// The order-by index refers to the position in the combined keys+data array
	// Keys come first, then data
	allValues := append(event.Keys, event.Data...)
	
	if idx.config.OrderBy < 0 || idx.config.OrderBy >= len(allValues) {
		// Invalid index, use block number as fallback
		return fmt.Sprintf("%020d", event.BlockNumber)
	}
	
	return padHex(allValues[idx.config.OrderBy])
}

// extractUniqueKey extracts the unique key value from the event
func (idx *Indexer) extractUniqueKey(event Event) string {
	// The unique index refers to the position in the combined keys+data array
	// Keys come first, then data
	allValues := append(event.Keys, event.Data...)
	
	if idx.config.Unique < 0 || idx.config.Unique >= len(allValues) {
		// Invalid index or unique constraint disabled
		return ""
	}
	
	return padHex(allValues[idx.config.Unique])
}
