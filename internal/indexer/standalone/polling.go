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
			
			// Index events from current block to latest (process a few blocks at a time)
			endBlock := latestBlock
			if endBlock > idx.currentBlock+10 {
				endBlock = idx.currentBlock + 10 // Process 10 blocks at a time
			}
			
			if idx.currentBlock <= endBlock {
				fmt.Printf("Processing blocks %d to %d (latest: %d)\n", idx.currentBlock, endBlock, latestBlock)
				
				for blockNum := idx.currentBlock; blockNum <= endBlock && idx.running; blockNum++ {
					events, err := idx.getEventsAtBlock(blockNum)
					if err != nil {
						fmt.Printf("Error getting events at block %d: %v\n", blockNum, err)
						continue
					}
					
					if len(events) > 0 {
						idx.storeEvents(events)
						fmt.Printf("Indexed %d events at block %d\n", len(events), blockNum)
						fmt.Printf("  Total events stored: %d\n", idx.GetEventCount())
					}
					
					idx.currentBlock = blockNum + 1
				}
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