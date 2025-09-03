package standalone

import (
	"fmt"
	"time"
)

// startPollingLoop starts the polling-based indexer loop
func (idx *Indexer) startPollingLoop() error {
	fmt.Printf("Starting polling indexer from block %d\n", idx.config.StartBlock)
	fmt.Printf("Indexing events from contract: %s\n", idx.config.Contract)
	fmt.Printf("Event: %s (selector: %s)\n", idx.config.Event, idx.eventSelector)
	
	// Main indexing loop
	for idx.running {
		select {
		case <-idx.stopChan:
			return nil
		default:
			// Get latest block number
			latestBlock, err := idx.getLatestBlockNumber()
			if err != nil {
				fmt.Printf("Error getting latest block: %v\n", err)
				time.Sleep(5 * time.Second)
				continue
			}
			
			// Index events from current block to latest (process larger ranges at a time)
			endBlock := latestBlock
			if endBlock > idx.currentBlock+100 {
				endBlock = idx.currentBlock + 100 // Process 100 blocks at a time
			}
			
			if idx.currentBlock <= endBlock {
				fmt.Printf("Processing blocks %d to %d (latest: %d)\n", idx.currentBlock, endBlock, latestBlock)
				
				// Process events in this range with continuation token support
				continuationToken := ""
				totalEventsInRange := 0
				
				for {
					events, nextToken, err := idx.getEventsInRange(idx.currentBlock, endBlock, continuationToken)
					if err != nil {
						fmt.Printf("Error getting events in range %d-%d: %v\n", idx.currentBlock, endBlock, err)
						break
					}
					
					if len(events) > 0 {
						if err := idx.storeEvents(events); err != nil {
							fmt.Printf("Error storing events: %v\n", err)
							break
						}
						totalEventsInRange += len(events)
						fmt.Printf("Indexed %d events (chunk), total in range: %d\n", len(events), totalEventsInRange)
					}
					
					// If no continuation token, we've processed all events in this range
					if nextToken == "" {
						break
					}
					
					continuationToken = nextToken
					fmt.Printf("Processing next chunk with continuation token...\n")
				}
				
				if totalEventsInRange > 0 {
					fmt.Printf("  Total events stored: %d\n", idx.GetEventCount())
				}
				
				// Move to next range
				idx.currentBlock = endBlock + 1
			}
			
			// Wait before next poll (shorter interval if we're catching up)
			if idx.currentBlock < latestBlock-50 {
				time.Sleep(100 * time.Millisecond) // Fast catchup
			} else {
				time.Sleep(2 * time.Second) // Normal polling
			}
		}
	}
	
	return nil
}