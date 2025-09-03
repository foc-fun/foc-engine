package standalone

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	DataDir    string // Directory for BadgerDB storage
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
	
	// Storage backend
	storage *Storage
	
	// State
	currentBlock  uint64
	running       bool
	stopChan      chan struct{}
	eventSelector string // Cached event selector
}

// determineStartingBlock determines the starting block for indexing
// Priority: max(lastProcessedBlock + 1, configStartBlock)
func determineStartingBlock(storage *Storage, configStartBlock uint64) (uint64, error) {
	lastProcessedBlock, err := storage.GetLastProcessedBlock()
	if err != nil {
		return 0, fmt.Errorf("failed to get last processed block: %v", err)
	}
	
	// If no blocks have been processed yet, use config start block
	if lastProcessedBlock == 0 {
		return configStartBlock, nil
	}
	
	// Resume from the block after the last processed block
	resumeBlock := lastProcessedBlock + 1
	
	// If config start block is ahead of our resume point, use config start block
	// This handles cases where user wants to skip ahead or restart from a later point
	if configStartBlock > resumeBlock {
		return configStartBlock, nil
	}
	
	return resumeBlock, nil
}

// New creates a new standalone indexer
func New(config Config) (*Indexer, error) {
	// Initialize storage
	dataDir := config.DataDir
	if dataDir == "" {
		dataDir = "./indexer_db" // Default fallback
	}
	storage, err := NewStorage(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}
	
	// Determine starting block: use last processed block or config start block
	startingBlock, err := determineStartingBlock(storage, config.StartBlock)
	if err != nil {
		storage.Close() // Clean up on error
		return nil, fmt.Errorf("failed to determine starting block: %v", err)
	}
	
	return &Indexer{
		config:       config,
		storage:      storage,
		currentBlock: startingBlock,
		stopChan:     make(chan struct{}),
	}, nil
}

// Start begins the indexing process, attempting WebSocket first with polling fallback
func (idx *Indexer) Start() error {
	idx.running = true
	
	// Compute and cache the event selector once
	idx.eventSelector = idx.computeEventSelector(idx.config.Event)
	
	// currentBlock is already set correctly in New() from determineStartingBlock()
	
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

// Close closes the indexer and storage
func (idx *Indexer) Close() error {
	if err := idx.Stop(); err != nil {
		return err
	}
	return idx.storage.Close()
}

// storeEvents stores events in BadgerDB
func (idx *Indexer) storeEvents(events []EventData) error {
	for _, event := range events {
		if err := idx.storage.StoreEvent(event); err != nil {
			return fmt.Errorf("failed to store event: %v", err)
		}
	}
	return nil
}

// GetEvents returns indexed events with pagination
func (idx *Indexer) GetEvents(page int, pageLength int, order string) ([]EventData, int, error) {
	return idx.storage.GetEvents(page, pageLength, order)
}

// GetEventCount returns the number of indexed events
func (idx *Indexer) GetEventCount() int {
	count, _ := idx.storage.GetEventCount()
	return count
}

// GetLastProcessedBlock returns the last processed block from storage
func (idx *Indexer) GetLastProcessedBlock() (uint64, error) {
	return idx.storage.GetLastProcessedBlock()
}

// GetCurrentBlock returns the current block the indexer will process next
func (idx *Indexer) GetCurrentBlock() uint64 {
	return idx.currentBlock
}

// GetLatestOrderedEvents returns the latest events ordered by the order key with unique constraint
func (idx *Indexer) GetLatestOrderedEvents(page int, pageLength int, order string) ([]EventData, int, error) {
	if idx.config.Unique < 0 {
		// No unique constraint, return all events
		return idx.storage.GetEvents(page, pageLength, order)
	}
	// Return unique events only
	return idx.storage.GetUniqueEvents(page, pageLength, order)
}

// GetUniqueEventCount returns the number of unique events
func (idx *Indexer) GetUniqueEventCount() int {
	count, _ := idx.storage.GetUniqueEventCount()
	return count
}

// parseQueryParams parses pagination and ordering parameters from the request
func parseQueryParams(r *http.Request) (page int, pageLength int, order string, err error) {
	// Parse page parameter (default: 0)
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		page = 0
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 0 {
			return 0, 0, "", fmt.Errorf("invalid page parameter: must be >= 0")
		}
	}
	
	// Parse pageLength parameter (default: 20)
	pageLengthStr := r.URL.Query().Get("pageLength")
	if pageLengthStr == "" {
		pageLength = 20
	} else {
		pageLength, err = strconv.Atoi(pageLengthStr)
		if err != nil || pageLength <= 0 || pageLength >= 100 {
			return 0, 0, "", fmt.Errorf("invalid pageLength parameter: must be > 0 and < 100")
		}
	}
	
	// Parse order parameter (default: asc)
	order = r.URL.Query().Get("order")
	if order == "" {
		order = "asc"
	} else if order != "asc" && order != "desc" {
		return 0, 0, "", fmt.Errorf("invalid order parameter: must be 'asc' or 'desc'")
	}
	
	return page, pageLength, order, nil
}


// startHTTPServer starts a simple HTTP server to query indexed data
func (idx *Indexer) startHTTPServer() {
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		page, pageLength, order, err := parseQueryParams(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		
		paginatedEvents, totalCount, err := idx.GetEvents(page, pageLength, order)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count": totalCount,
			"page":        page,
			"page_length": pageLength,
			"order":       order,
			"count":       len(paginatedEvents),
			"events":      paginatedEvents,
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
		// Parse query parameters
		page, pageLength, order, err := parseQueryParams(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		
		paginatedEvents, totalCount, err := idx.GetLatestOrderedEvents(page, pageLength, order)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count":         totalCount,
			"page":                page,
			"page_length":         pageLength,
			"order":               order,
			"count":               len(paginatedEvents),
			"unique_enabled":      idx.config.Unique >= 0,
			"order_by_index":      idx.config.OrderBy,
			"unique_key_index":    idx.config.Unique,
			"events":              paginatedEvents,
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

// padHex pads a hex string to 32 bytes (66 characters including 0x prefix)
func padHex(hex string) string {
	if hex == "" {
		return "0x0000000000000000000000000000000000000000000000000000000000000000"
	}
	
	// Remove 0x prefix if present
	if strings.HasPrefix(hex, "0x") {
		hex = hex[2:]
	}
	
	// Pad to 64 characters (32 bytes)
	if len(hex) < 64 {
		hex = strings.Repeat("0", 64-len(hex)) + hex
	}
	
	return "0x" + hex
}

// padHexArray pads all hex strings in an array to 32 bytes
func padHexArray(hexArray []string) []string {
	paddedArray := make([]string, len(hexArray))
	for i, hex := range hexArray {
		paddedArray[i] = padHex(hex)
	}
	return paddedArray
}

// computeEventSelector computes the event selector from event name
func (idx *Indexer) computeEventSelector(eventName string) string {
	// Delegate to the RPC module which has the logic
	return idx.getEventSelector(eventName)
}