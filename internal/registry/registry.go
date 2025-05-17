package registry

import (
	"fmt"

	"github.com/b-j-roberts/foc-engine/internal/provider"
)

type RegisteredContract struct {
  Address string
  ClassHash string
  ContractClass *provider.ContractClass
}

type RegisteredClass struct {
  Address string
  Name    string
  Version string
}

type Registry struct {
  RegistryAddresses map[string]bool
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
  FocRegistry.RegistryAddresses[address] = true
}

func RegisterContract(address string, classHash string) {
  if FocRegistry == nil {
    FocRegistry = &Registry{}
  }
  if FocRegistry.RegisteredContracts == nil {
    FocRegistry.RegisteredContracts = make(map[string]RegisteredContract)
  }
  contractClass, err := provider.GetStarknetClassAt(address)
  if err != nil {
    fmt.Println("Error getting contract class:", err)
    return
  }

  FocRegistry.RegisteredContracts[address] = RegisteredContract{
    Address: address,
    ClassHash: classHash,
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
