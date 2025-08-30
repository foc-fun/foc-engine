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
	Unique     int // Key index for unique constraint (-1 to disable)
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
	OrderKey        string   `json:"order_key"`  // The key value used for ordering
	UniqueKey       string   `json:"unique_key"` // The key value used for uniqueness constraint
}

// Indexer handles event indexing from Starknet
type Indexer struct {
	config Config
	
	// In-memory storage
	events       []EventData                // All events (for backward compatibility)
	uniqueEvents map[string]EventData       // Latest events by unique key
	eventsMux    sync.RWMutex
	
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
		uniqueEvents: make(map[string]EventData),
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
	
	for _, event := range events {
		// Store in all events list (for backward compatibility)
		idx.events = append(idx.events, event)
		
		// Store in unique events map if unique constraint is enabled
		if idx.config.Unique >= 0 && event.UniqueKey != "" {
			idx.uniqueEvents[event.UniqueKey] = event
		}
	}
	
	// Sort all events by the order key
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

// GetLatestOrderedEvents returns the latest events ordered by the order key with unique constraint
func (idx *Indexer) GetLatestOrderedEvents() []EventData {
	idx.eventsMux.RLock()
	defer idx.eventsMux.RUnlock()
	
	if idx.config.Unique < 0 {
		// No unique constraint, return all events ordered
		eventsCopy := make([]EventData, len(idx.events))
		copy(eventsCopy, idx.events)
		return eventsCopy
	}
	
	// Extract values from unique events map and sort them
	events := make([]EventData, 0, len(idx.uniqueEvents))
	for _, event := range idx.uniqueEvents {
		events = append(events, event)
	}
	
	// Sort by the order key
	sort.Slice(events, func(i, j int) bool {
		return events[i].OrderKey < events[j].OrderKey
	})
	
	return events
}

// GetUniqueEventCount returns the number of unique events
func (idx *Indexer) GetUniqueEventCount() int {
	idx.eventsMux.RLock()
	defer idx.eventsMux.RUnlock()
	return len(idx.uniqueEvents)
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
			"running":             idx.running,
			"current_block":       idx.currentBlock,
			"event_count":         idx.GetEventCount(),
			"unique_event_count":  idx.GetUniqueEventCount(),
			"unique_enabled":      idx.config.Unique >= 0,
			"config":              idx.config,
		})
	})
	
	http.HandleFunc("/events-latest-ordered", func(w http.ResponseWriter, r *http.Request) {
		events := idx.GetLatestOrderedEvents()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":               len(events),
			"unique_enabled":      idx.config.Unique >= 0,
			"order_by_index":      idx.config.OrderBy,
			"unique_key_index":    idx.config.Unique,
			"events":              events,
		})
	})
	
	port := 8080
	fmt.Printf("HTTP server listening on :%d\n", port)
	fmt.Println("Query endpoints:")
	fmt.Printf("  http://localhost:%d/status - Get indexer status\n", port)
	fmt.Printf("  http://localhost:%d/events - Get all indexed events\n", port)
	fmt.Printf("  http://localhost:%d/events-latest-ordered - Get latest events with unique constraint\n", port)
	
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Printf("HTTP server error: %v\n", err)
	}
}

// computeEventSelector computes the event selector from event name
func (idx *Indexer) computeEventSelector(eventName string) string {
	// Delegate to the RPC module which has the logic
	return idx.getEventSelector(eventName)
}