package registry

import (
	"fmt"

	"github.com/NethermindEth/starknet.go/contracts"

	"github.com/b-j-roberts/foc-engine/internal/provider"
)

type RegisteredContract struct {
	Address                 string
	ClassHash               string
	ContractClass           *provider.ContractClass
	NethermindContractClass *contracts.ContractClass
}

type RegisteredClass struct {
	Address string
	Name    string
	Version string
}

type Registry struct {
	// Map: RegistryAddress -> isRegistered
	RegistryAddresses  map[string]bool
	RegistryContracts  map[string]RegisteredContract
	LastCompletedBlock uint
	// Map: ContractAddress -> RegisteredContract
	RegisteredContracts map[string]RegisteredContract
}

var FocRegistry *Registry

func AddRegistryAddress(address string) {
	if FocRegistry == nil {
		FocRegistry = &Registry{}
	}
	if FocRegistry.RegistryAddresses == nil {
		FocRegistry.RegistryAddresses = make(map[string]bool)
	}
	contractAddress := address
	fmt.Println("Adding registry address:", contractAddress)
	if len(contractAddress) != 66 {
		// Remove 0x prefix if present
		if contractAddress[:2] == "0x" {
			contractAddress = contractAddress[2:]
		}
		// Pad with leading zeros to 64 characters
		contractAddress = fmt.Sprintf("0x%064s", contractAddress)
	}
	FocRegistry.RegistryAddresses[contractAddress] = true

	contractClass, err := provider.GetStarknetClassAt(contractAddress)
	if err != nil {
		fmt.Println("Error getting contract class:", err)
		return
	}
	if FocRegistry.RegistryContracts == nil {
		FocRegistry.RegistryContracts = make(map[string]RegisteredContract)
	}
	FocRegistry.RegistryContracts[contractAddress] = RegisteredContract{
		Address:       contractAddress,
		ClassHash:     "0x0", // TODO
		ContractClass: contractClass,
	}
}

func RegisterContract(address string, classHash string) {
	if FocRegistry == nil {
		FocRegistry = &Registry{}
	}
	if FocRegistry.RegisteredContracts == nil {
		FocRegistry.RegisteredContracts = make(map[string]RegisteredContract)
	}
	contractAddress := address
	if len(contractAddress) != 66 {
		// Remove 0x prefix if present
		if contractAddress[:2] == "0x" {
			contractAddress = contractAddress[2:]
		}
		// Pad with leading zeros to 64 characters
		contractAddress = fmt.Sprintf("0x%064s", contractAddress)
	}
	contractClass, err := provider.GetStarknetClassAt(contractAddress)
	if err != nil {
		fmt.Println("Error getting contract class:", err)
		return
	}

	FocRegistry.RegisteredContracts[contractAddress] = RegisteredContract{
		Address:       contractAddress,
		ClassHash:     classHash,
		ContractClass: contractClass,
	}
}

func RegisterClass(address string, name string, version string) {
	if FocRegistry == nil {
		FocRegistry = &Registry{}
	}
	/*
	  TODO
	  if FocRegistry.RegisteredContracts == nil {
	    FocRegistry.RegisteredContracts = make(map[string]RegisteredContract)
	  }
	  FocRegistry.RegisteredContracts[address] = RegisteredContract{
	    Address: address,
	    Name:    name,
	    Version: version,
	  }
	*/
}
