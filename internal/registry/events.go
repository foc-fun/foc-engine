package registry

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/b-j-roberts/foc-engine/internal/db/mongo"
	"github.com/b-j-roberts/foc-engine/internal/provider"
)

// TODO: Hardcoded
const (
	ClassRegisteredEvent    = "0x11fcf2735cfadde3c48253da6e8eacdf6030dc3694cc3d710b22214d6a2ed19"
	ContractRegisteredEvent = "0x206ba27d5bbda42a63e108ee1ac7a6455c197ee34cd40a268e61b06f78dbc9a"
)

type StarknetEventData struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		SubscriptionId big.Int `json:"subscription_id"`
		Result         struct {
			BlockHash       string   `json:"block_hash"`
			BlockNumber     uint     `json:"block_number"`
			FromAddress     string   `json:"from_address"`
			TransactionHash string   `json:"transaction_hash"`
			Keys            []string `json:"keys"`
			Data            []string `json:"data"`
		}
	} `json:"params"`
}

func PrintStarknetEventData(eventData StarknetEventData) {
	fmt.Println("Received event data:")
	fmt.Printf("  Subscription ID: %s\n", eventData.Params.SubscriptionId.String())
	fmt.Printf("  Block Hash: %s\n", eventData.Params.Result.BlockHash)
	fmt.Printf("  Block Number: %d\n", eventData.Params.Result.BlockNumber)
	fmt.Printf("  From Address: %s\n", eventData.Params.Result.FromAddress)
	fmt.Printf("  Transaction Hash: %s\n", eventData.Params.Result.TransactionHash)
	fmt.Printf("  Keys: %v\n", eventData.Params.Result.Keys)
	fmt.Printf("  Data: %v\n", eventData.Params.Result.Data)
}

func ProcessStarknetEventData(message []byte) {
	var eventData StarknetEventData
	err := json.Unmarshal(message, &eventData)
	if err != nil {
		fmt.Println("Error unmarshalling event data:", err)
		return
	}
	// TODO: Pad the address to 0x0000...0000 w/ 64 hex digits
	contractAddress := eventData.Params.Result.FromAddress
	if len(contractAddress) != 66 {
		// Remove 0x prefix if present
		if contractAddress[:2] == "0x" {
			contractAddress = contractAddress[2:]
		}
		// Pad with leading zeros to 64 characters
		contractAddress = fmt.Sprintf("0x%064s", contractAddress)
	}
	if FocRegistry.RegistryAddresses[contractAddress] {
		ProcessRegistryEvent(eventData)
	} else if _, ok := FocRegistry.RegisteredContracts[contractAddress]; ok {
		ProcessRegisteredContractEvent(eventData)
	} else {
		fmt.Println("Unknown contract address:", contractAddress)
	}
}

// TODO: Store current block events locally till block complete?
func ProcessRegistryEvent(eventMessage StarknetEventData) {
	if eventMessage.Method == "" {
		return
	}

	if eventMessage.Params.Result.Keys[0] == ContractRegisteredEvent {
		ProcessRegisterContractEvent(eventMessage)
	} else {
		fmt.Println("Unknown event:")
		PrintStarknetEventData(eventMessage)
	}

	// Track the last completed blocks
	lastCompletedBlock := FocRegistry.LastCompletedBlock
	// One-off offset to ensure we don't miss any events if shut down mid-block
	if eventMessage.Params.Result.BlockNumber > lastCompletedBlock+1 {
		FocRegistry.LastCompletedBlock = eventMessage.Params.Result.BlockNumber - 1
	}
}

func ProcessRegisterContractEvent(eventMessage StarknetEventData) {
  focEngineAddress := eventMessage.Params.Result.FromAddress
  if len(focEngineAddress) != 66 {
    // Remove 0x prefix if present
    if focEngineAddress[:2] == "0x" {
      focEngineAddress = focEngineAddress[2:]
    }
    // Pad with leading zeros to 64 characters
    focEngineAddress = fmt.Sprintf("0x%064s", focEngineAddress)
  }
	address := eventMessage.Params.Result.Keys[1]
	classHash := eventMessage.Params.Result.Data[0]
	RegisterContract(address, classHash)

  if _, ok := FocRegistry.RegistryContracts[focEngineAddress]; !ok {
    fmt.Println("Unknown foc engine address:", focEngineAddress)
    return
  }
  registryContract := FocRegistry.RegistryContracts[focEngineAddress]
  abi := registryContract.ContractClass.Abi
  typeName, err := GetEventTypeName(eventMessage.Params.Result.Keys[0], abi)
  if err != nil {
    fmt.Println("Error getting event type name:", err)
    return
  }
  eventData := eventMessage.Params.Result.Keys[1:]
  eventData = append(eventData, eventMessage.Params.Result.Data...)
  typeNameJson, _ := StarknetTypeDataMin(typeName, abi, eventData)
  typeNameJson.(map[string]interface{})["registry_address"] = eventMessage.Params.Result.FromAddress
  typeNameJson.(map[string]interface{})["block_number"] = eventMessage.Params.Result.BlockNumber
  typeNameJson.(map[string]interface{})["transaction_hash"] = eventMessage.Params.Result.TransactionHash
  typeNameJson.(map[string]interface{})["event_type"] = typeName
	res, err := mongo.InsertJson("foc_engine", "registry", typeNameJson)
	if err != nil {
		fmt.Println("Error inserting event into MongoDB:", err)
		return
	}
	fmt.Println("Inserted event into MongoDB:", res)

	provider.SubscribeEvents(address)
	fmt.Println("Subscribed to events for contract:", address)
}

func ProcessRegisterClassEvent(eventMessage StarknetEventData) {
	// Register contract in memory
	address := eventMessage.Params.Result.Keys[1]
	name := eventMessage.Params.Result.Data[0]
	version := eventMessage.Params.Result.Data[1]
	RegisterClass(address, name, version)

	// TODO: Load abi

	// TODO: Insert only essential data into Mongo
	// Insert Registry into Mongo for use on restart
	res, err := mongo.InsertJson("foc_engine", "registry", eventMessage.Params.Result)
	if err != nil {
		fmt.Println("Error inserting event into MongoDB:", err)
		return
	}
	fmt.Println("Inserted event into MongoDB:", res)
}

func ProcessRegisteredContractEvent(eventMessage StarknetEventData) {
	contractAddress := eventMessage.Params.Result.FromAddress
	if len(contractAddress) != 66 {
		// Remove 0x prefix if present
		if contractAddress[:2] == "0x" {
			contractAddress = contractAddress[2:]
		}
		// Pad with leading zeros to 64 characters
		contractAddress = fmt.Sprintf("0x%064s", contractAddress)
	}
	if _, ok := FocRegistry.RegisteredContracts[contractAddress]; !ok {
		fmt.Println("Unknown registered contract address:", contractAddress)
		return
	}
	registeredContract := FocRegistry.RegisteredContracts[contractAddress]
	abi := registeredContract.ContractClass.Abi
	typeName, err := GetEventTypeName(eventMessage.Params.Result.Keys[0], abi)
	if err != nil {
		fmt.Println("Error getting event type name:", err)
		return
	}
	eventData := eventMessage.Params.Result.Keys[1:]
	eventData = append(eventData, eventMessage.Params.Result.Data...)
	typeNameJson, _ := StarknetTypeDataMin(typeName, abi, eventData)
	typeNameJson.(map[string]interface{})["contract_address"] = eventMessage.Params.Result.FromAddress
	typeNameJson.(map[string]interface{})["block_number"] = eventMessage.Params.Result.BlockNumber
	typeNameJson.(map[string]interface{})["transaction_hash"] = eventMessage.Params.Result.TransactionHash
	typeNameJson.(map[string]interface{})["event_type"] = typeName

	_, err = mongo.InsertJson("foc_engine", "events", typeNameJson)
	if err != nil {
		fmt.Println("Error inserting event into MongoDB:", err)
		return
	}
}
