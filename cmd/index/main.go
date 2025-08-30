package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/b-j-roberts/foc-engine/internal/indexer/standalone"
)

func main() {
	// Define command-line flags
	contract := flag.String("contract", "", "Contract address to index events from (required)")
	event := flag.String("event", "", "Event name to index (required)")
	orderBy := flag.Int("order-by", 0, "Key index to order events by (required)")
	unique := flag.Int("unique", -1, "Key index for unique constraint (-1 to disable)")
	startBlock := flag.Uint64("start-block", 0, "Starting block number (required)")
	rpc := flag.String("rpc", "", "RPC endpoint URL (required)")
	network := flag.String("network", "sepolia", "Network to connect to (devnet, sepolia, mainnet)")
	
	flag.Parse()
	
	// Validate required flags
	if *contract == "" {
		fmt.Println("Error: --contract flag is required")
		flag.Usage()
		os.Exit(1)
	}
	if *event == "" {
		fmt.Println("Error: --event flag is required")
		flag.Usage()
		os.Exit(1)
	}
	if *rpc == "" {
		fmt.Println("Error: --rpc flag is required")
		flag.Usage()
		os.Exit(1)
	}
	
	// Print indexer configuration
	fmt.Println("Starting FOC Indexer with configuration:")
	fmt.Printf("  Contract: %s\n", *contract)
	fmt.Printf("  Event: %s\n", *event)
	fmt.Printf("  Order By Key: %d\n", *orderBy)
	fmt.Printf("  Start Block: %d\n", *startBlock)
	fmt.Printf("  RPC: %s\n", *rpc)
	fmt.Printf("  Network: %s\n", *network)
	fmt.Println()
	
	// Create and start the indexer
	indexer, err := standalone.New(standalone.Config{
		Contract:   *contract,
		Event:      *event,
		OrderBy:    *orderBy,
		Unique:     *unique,
		StartBlock: *startBlock,
		RPC:        *rpc,
		Network:    *network,
	})
	if err != nil {
		fmt.Printf("Error creating indexer: %v\n", err)
		os.Exit(1)
	}
	
	// Start indexing in a goroutine
	go func() {
		if err := indexer.Start(); err != nil {
			fmt.Printf("Error starting indexer: %v\n", err)
			os.Exit(1)
		}
	}()
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	<-sigChan
	fmt.Println("\nShutting down indexer...")
	
	if err := indexer.Close(); err != nil {
		fmt.Printf("Error closing indexer: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Indexer stopped successfully")
}