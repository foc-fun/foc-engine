package standalone

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
)

// Config holds the configuration for the standalone indexer
type Config struct {
	Contract   string
	Event      string
	OrderBy    int
	StartBlock uint64
	RPC        string
	Network    string
}

// EventData represents a single indexed event
type EventData struct {
	BlockNumber     uint64   `json:"block_number"`
	TransactionHash string   `json:"transaction_hash"`
	FromAddress     string   `json:"from_address"`
	Keys            []string `json:"keys"`
	Data            []string `json:"data"`
	Timestamp       int64    `json:"timestamp"`
	OrderKey        string   `json:"order_key"` // The key value used for ordering
}

// Indexer handles event indexing from Starknet
type Indexer struct {
	config Config
	
	// In-memory storage
	events    []EventData
	eventsMux sync.RWMutex
	
	// State
	currentBlock  uint64
	running       bool
	stopChan      chan struct{}
	eventSelector string // Cached event selector
}

// New creates a new standalone indexer
func New(config Config) *Indexer {
	return &Indexer{
		config:       config,
		events:       make([]EventData, 0),
		currentBlock: config.StartBlock,
		stopChan:     make(chan struct{}),
	}
}

// Start begins the indexing process, attempting WebSocket first with polling fallback
func (idx *Indexer) Start() error {
	idx.running = true
	
	// Compute and cache the event selector once
	idx.eventSelector = idx.computeEventSelector(idx.config.Event)
	
	// Set current block to start block
	idx.currentBlock = idx.config.StartBlock
	
	// Start HTTP server for querying indexed data (only once)
	go idx.startHTTPServer()
	
	// Try WebSocket first
	fmt.Println("Attempting WebSocket connection...")
	err := idx.tryWebSocket()
	
	// If WebSocket fails, fall back to polling
	if err != nil {
		fmt.Printf("WebSocket failed (%v), falling back to polling mode...\n", err)
		return idx.startPollingLoop()
	}
	
	return nil
}

// Stop gracefully stops the indexer
func (idx *Indexer) Stop() error {
	idx.running = false
	close(idx.stopChan)
	return nil
}

// storeEvents stores events in memory, maintaining order by the specified key
func (idx *Indexer) storeEvents(events []EventData) {
	idx.eventsMux.Lock()
	defer idx.eventsMux.Unlock()
	
	idx.events = append(idx.events, events...)
	
	// Sort by the order key
	sort.Slice(idx.events, func(i, j int) bool {
		return idx.events[i].OrderKey < idx.events[j].OrderKey
	})
}

// GetEvents returns all indexed events
func (idx *Indexer) GetEvents() []EventData {
	idx.eventsMux.RLock()
	defer idx.eventsMux.RUnlock()
	
	// Return a copy to prevent external modification
	eventsCopy := make([]EventData, len(idx.events))
	copy(eventsCopy, idx.events)
	return eventsCopy
}

// GetEventCount returns the number of indexed events
func (idx *Indexer) GetEventCount() int {
	idx.eventsMux.RLock()
	defer idx.eventsMux.RUnlock()
	return len(idx.events)
}

// startHTTPServer starts a simple HTTP server to query indexed data
func (idx *Indexer) startHTTPServer() {
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		events := idx.GetEvents()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":  len(events),
			"events": events,
		})
	})
	
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"running":       idx.running,
			"current_block": idx.currentBlock,
			"event_count":   idx.GetEventCount(),
			"config":        idx.config,
		})
	})
	
	port := 8080
	fmt.Printf("HTTP server listening on :%d\n", port)
	fmt.Println("Query endpoints:")
	fmt.Printf("  http://localhost:%d/status - Get indexer status\n", port)
	fmt.Printf("  http://localhost:%d/events - Get all indexed events\n", port)
	
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Printf("HTTP server error: %v\n", err)
	}
}

// computeEventSelector computes the event selector from event name
func (idx *Indexer) computeEventSelector(eventName string) string {
	// Delegate to the RPC module which has the logic
	return idx.getEventSelector(eventName)
}