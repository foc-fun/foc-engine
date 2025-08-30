package standalone

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// Storage handles persistent storage of events using BadgerDB
type Storage struct {
	db *badger.DB
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string) (*Storage, error) {
	// Create data directory if it doesn't exist
	dbPath := filepath.Join(dataDir, "indexer_data")
	
	// Open BadgerDB
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable BadgerDB logs for now
	
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %v", err)
	}
	
	return &Storage{db: db}, nil
}

// Close closes the storage
func (s *Storage) Close() error {
	return s.db.Close()
}

// StoreEvent stores an event with multiple indexes
func (s *Storage) StoreEvent(event EventData) error {
	return s.db.Update(func(txn *badger.Txn) error {
		// Serialize event data
		eventJSON, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %v", err)
		}
		
		// Primary key: order_key + block_number + tx_hash
		// This ensures uniqueness and maintains order
		primaryKey := fmt.Sprintf("event:order:%s:%020d:%s", 
			event.OrderKey, 
			event.BlockNumber, 
			event.TransactionHash)
		
		// Store the event
		if err := txn.Set([]byte(primaryKey), eventJSON); err != nil {
			return fmt.Errorf("failed to store event: %v", err)
		}
		
		// If unique key is set, store/update the unique index
		if event.UniqueKey != "" {
			uniqueKey := fmt.Sprintf("event:unique:%s", event.UniqueKey)
			if err := txn.Set([]byte(uniqueKey), eventJSON); err != nil {
				return fmt.Errorf("failed to store unique index: %v", err)
			}
		}
		
		// Store a reverse index for descending queries
		reverseKey := fmt.Sprintf("event:reverse:%s:%020d:%s",
			reverseString(event.OrderKey),
			999999999999-event.BlockNumber,
			event.TransactionHash)
		if err := txn.Set([]byte(reverseKey), eventJSON); err != nil {
			return fmt.Errorf("failed to store reverse index: %v", err)
		}
		
		return nil
	})
}

// GetEvents retrieves events with pagination and ordering
func (s *Storage) GetEvents(page int, pageLength int, order string) ([]EventData, int, error) {
	var events []EventData
	var totalCount int
	
	err := s.db.View(func(txn *badger.Txn) error {
		// Determine prefix based on order
		prefix := "event:order:"
		if order == "desc" {
			prefix = "event:reverse:"
		}
		
		// First, count total events
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		
		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			totalCount++
		}
		
		// Now fetch the requested page
		it2 := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it2.Close()
		
		startIdx := page * pageLength
		currentIdx := 0
		
		for it2.Seek(prefixBytes); it2.ValidForPrefix(prefixBytes); it2.Next() {
			// Skip to the start of the page
			if currentIdx < startIdx {
				currentIdx++
				continue
			}
			
			// Stop if we've collected enough events
			if len(events) >= pageLength {
				break
			}
			
			item := it2.Item()
			err := item.Value(func(val []byte) error {
				var event EventData
				if err := json.Unmarshal(val, &event); err != nil {
					return err
				}
				events = append(events, event)
				return nil
			})
			if err != nil {
				return err
			}
			
			currentIdx++
		}
		
		return nil
	})
	
	return events, totalCount, err
}

// GetUniqueEvents retrieves unique events with pagination and ordering
func (s *Storage) GetUniqueEvents(page int, pageLength int, order string) ([]EventData, int, error) {
	// First, get all unique events
	uniqueEvents := make(map[string]EventData)
	
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		
		prefix := []byte("event:unique:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.Key())
			uniqueKey := strings.TrimPrefix(key, "event:unique:")
			
			err := item.Value(func(val []byte) error {
				var event EventData
				if err := json.Unmarshal(val, &event); err != nil {
					return err
				}
				uniqueEvents[uniqueKey] = event
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	
	if err != nil {
		return nil, 0, err
	}
	
	// Convert map to slice for sorting
	allEvents := make([]EventData, 0, len(uniqueEvents))
	for _, event := range uniqueEvents {
		allEvents = append(allEvents, event)
	}
	
	// Sort events by OrderKey
	sortEventsByOrderKey(allEvents, order == "desc")
	
	// Apply pagination
	totalCount := len(allEvents)
	start := page * pageLength
	if start >= totalCount {
		return []EventData{}, totalCount, nil
	}
	
	end := start + pageLength
	if end > totalCount {
		end = totalCount
	}
	
	return allEvents[start:end], totalCount, nil
}

// GetEventCount returns the total number of events
func (s *Storage) GetEventCount() (int, error) {
	count := 0
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		
		prefix := []byte("event:order:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			count++
		}
		return nil
	})
	return count, err
}

// GetUniqueEventCount returns the number of unique events
func (s *Storage) GetUniqueEventCount() (int, error) {
	count := 0
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		
		prefix := []byte("event:unique:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			count++
		}
		return nil
	})
	return count, err
}

// ClearAll removes all data from storage (useful for testing)
func (s *Storage) ClearAll() error {
	return s.db.DropAll()
}

// reverseString reverses a string for reverse ordering
func reverseString(s string) string {
	// For hex strings, we can invert for reverse ordering
	// This is a simple approach - could be optimized
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// sortEventsByOrderKey sorts events by their OrderKey
func sortEventsByOrderKey(events []EventData, descending bool) {
	// Use a simple sort for now - could be optimized with a better algorithm
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			shouldSwap := false
			if descending {
				shouldSwap = events[i].OrderKey < events[j].OrderKey
			} else {
				shouldSwap = events[i].OrderKey > events[j].OrderKey
			}
			
			if shouldSwap {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}