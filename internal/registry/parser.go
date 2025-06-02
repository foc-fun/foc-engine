package registry

import (
	"fmt"
	"strconv"

	"github.com/NethermindEth/starknet.go/contracts"
	"github.com/NethermindEth/starknet.go/utils"
)

func GetEventTypeName(eventSelector string, abi []interface{}) (string, error) {
	if abi == nil {
		return "", fmt.Errorf("abi is nil")
	}
	/*
	  var nestedAbi contracts.NestedString
	  if err := nestedAbi.UnmarshalJSON([]byte(abi)); err != nil {
	    return "", fmt.Errorf("failed to unmarshal abi: %v", err)
	  }
	  fmt.Println("Nested ABI:", nestedAbi)
	*/
	for _, abiEntry := range abi {
		checkABI, ok := abiEntry.(map[string]interface{})
		if !ok {
			continue
		}
		abiType, ok := checkABI["type"].(string)
		if !ok {
			continue
		}
		switch abiType {
		case string(contracts.ABITypeConstructor), string(contracts.ABITypeFunction), string(contracts.ABITypeL1Handler):
			//TODO: fmt.Println("ABI Function Entry:", abiEntry)
			continue
		case string(contracts.ABITypeStruct):
			// TODO: fmt.Println("ABI Struct Entry:", abiEntry)
			continue
		case string(contracts.ABITypeEvent):
			// Example output for event types
			/*
			   indexer-1          | ABI Event Entry: map[kind:enum name:pow_game::pow::PowGame::Event type:event variants:[map[kind:nested name:ChainUnlocked type:pow_game::pow::PowGame::ChainUnlocked] map[kind:nested name:BalanceUpdated type:pow_game::pow::PowGame::BalanceUpdated] map[kind:nested name:TransactionAdded type:pow_game::actions::TransactionAdded] map[kind:nested name:BlockMined type:pow_game::actions::BlockMined] map[kind:nested name:DAStored type:pow_game::actions::DAStored] map[kind:nested name:ProofStored type:pow_game::actions::ProofStored] map[kind:flat name:UpgradeEvent type:pow_game::upgrades::component::PowUpgradesComponent::Event] map[kind:flat name:TransactionEvent type:pow_game::transactions::component::PowTransactionsComponent::Event] map[kind:flat name:PrestigeEvent type:pow_game::prestige::component::PrestigeComponent::Event] map[kind:flat name:BuilderEvent type:pow_game::builder::component::BuilderComponent::Event]]]

			*/
			// If event entry ends with "<ContractName::EventName>" ending
			//if stringEndsWith(checkABI["name"].(string), "::"+ContractName+"::Event") {
			if stringEndsWith(checkABI["name"].(string), "::Event") {
				// Loop through variants
				for _, variant := range checkABI["variants"].([]interface{}) {
					variantEntry, ok := variant.(map[string]interface{})
					if !ok {
						continue
					}
					variantName, ok := variantEntry["name"].(string)
					if !ok {
						continue
					}
					// Check if variant name is the same as event selector
					if utils.GetSelectorFromNameFelt(variantName).String() == eventSelector {
						variantType, ok := variantEntry["type"].(string)
						if !ok {
							continue
						}
						// Remove prefix from variantType
						// Example: "pow_game::pow::PowGame::ChainUnlocked" -> "ChainUnlocked"
						//variantType = variantType[strings.LastIndex(variantType, "::")+2:]
						return variantType, nil
					}
				}
			}
		default:
			// fmt.Println("Unknown ABI Entry:", abiType)
			continue
		}
	}

	return "", fmt.Errorf("event not found")
}

// TODO: Check valid data
// TODO: Parse data based on type
func StarknetTypeDataMin(typeName string, abis []interface{}, data []string) (interface{}, int) {
	if IsPrimitiveType(typeName) {
		return StarknetStringToTypedData(typeName, data[0]), 1
	} else if IsArrayType(typeName) {
		arrayType := GetArrayInnerType(typeName)
		arrayLen, err := strconv.ParseUint(data[0], 0, 32)
		data = data[1:]
		if err != nil {
			fmt.Println("Error parsing array length:", err)
			return nil, 0
		}
		var arrayData []interface{}
		totalOffset := 1
		for i := 0; i < int(arrayLen); i++ {
			value, offset := StarknetTypeDataMin(arrayType, abis, data)
			arrayData = append(arrayData, value)
			data = data[offset:]
			totalOffset += offset
		}
		return arrayData, totalOffset
	} else if IsStructType(typeName) {
		fields := map[string]interface{}{}
		totalOffset := 0
		for _, abi := range abis {
			if abi.(map[string]interface{})["name"] == typeName ||
				stringEndsWith(abi.(map[string]interface{})["name"].(string), "::"+typeName) {
				for _, member := range abi.(map[string]interface{})["members"].([]interface{}) {
					memberName := member.(map[string]interface{})["name"]
					value, offset := StarknetTypeDataMin(member.(map[string]interface{})["type"].(string), abis, data)
					fields[memberName.(string)] = value
					data = data[offset:]
					totalOffset += offset
				}
				break
			}
		}
		return fields, totalOffset
	}
	return nil, 0
}

func StarknetStringToTypedData(typeName string, data string) interface{} {
	if IsPrimitiveType(typeName) {
		return StarknetTypeParsers[typeName](typeName, data)
	} else {
		fmt.Println("Not a primitive type")
		// TODO: Error?
	}
	return nil
}

func IsArrayType(typeName string) bool {
	for _, array := range Types.Array {
		if typeName == array.Type ||
			stringStartsWith(typeName, array.Type+"<") ||
			stringStartsWith(typeName, array.Type+"::<") {
			return true
		}
	}
	return false
}

func GetArrayInnerType(typeName string) string {
	for _, array := range Types.Array {
		if stringStartsWith(typeName, array.Type+"<") {
			start := len(array.Type) + 1
			end := len(typeName) - 1
			return typeName[start:end]
		} else if stringStartsWith(typeName, array.Type+"::<") {
			start := len(array.Type) + 3
			end := len(typeName) - 1
			return typeName[start:end]
		} else if typeName == array.Type {
			return array.Type
		}
	}
	return ""
}

func IsStructType(typeName string) bool {
	return !IsPrimitiveType(typeName) && !IsArrayType(typeName)
}

func IsPrimitiveType(typeName string) bool {
	for _, primitive := range Types.Primitives {
		if primitive.Type == typeName {
			return true
		}
	}
	return false
}

type TypeInfo struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type TypesInfo struct {
	Primitives []TypeInfo `json:"primitives"`
	Array      []TypeInfo `json:"array"`
}

var Types = TypesInfo{
	Primitives: []TypeInfo{
		{
			Type: "core::byte_array::ByteArray",
			Name: "string",
		},
		{
			Type: "core::felt252",
			Name: "felt",
		},
		{
			Type: "core::integer::u8",
			Name: "u8",
		},
		{
			Type: "core::integer::u16",
			Name: "u16",
		},
		{
			Type: "core::integer::u32",
			Name: "u32",
		},
		{
			Type: "core::integer::u64",
			Name: "u64",
		},
		{
			Type: "core::integer::u128",
			Name: "u128",
		},
		{
			Type: "core::integer::u256",
			Name: "u256",
		},
		{
			Type: "core::integer::i8",
			Name: "i8",
		},
		{
			Type: "core::integer::i16",
			Name: "i16",
		},
		{
			Type: "core::integer::i32",
			Name: "i32",
		},
		{
			Type: "core::integer::i64",
			Name: "i64",
		},
		{
			Type: "core::integer::i128",
			Name: "i128",
		},
		{
			Type: "core::bool",
			Name: "bool",
		},
		{
			Type: "core::starknet::contract_address::ContractAddress",
			Name: "address",
		},
		{
			Type: "core::starknet::class_hash::ClassHash",
			Name: "class_hash",
		},
	},
	Array: []TypeInfo{
		{
			Type: "core::array::Array",
			Name: "array",
		},
		{
			Type: "core::span::Span",
			Name: "span",
		},
		{
			Type: "core::array::Span",
			Name: "span",
		},
		{
			Type: "@core::array::Array",
			Name: "array",
		},
		{
			Type: "@core::span::Span",
			Name: "span",
		},
		{
			Type: "@core::array::Span",
			Name: "span",
		},
	},
}

// TODO: Improve "snapshot" types

func stringStartsWith(s string, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
func stringEndsWith(s string, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

var StarknetTypeParsers = map[string]func(string, string) interface{}{
	"core::byte_array::ByteArray": func(typeName string, data string) interface{} {
		// TODO: Parse byte array
		return data
	},
	"core::felt252": func(typeName string, data string) interface{} {
		return data
	},
	"core::integer::u8": func(typeName string, data string) interface{} {
		val, err := strconv.ParseUint(data, 0, 8)
		if err != nil {
			return nil
		}
		return val
	},
	"core::integer::u16": func(typeName string, data string) interface{} {
		val, err := strconv.ParseUint(data, 0, 16)
		if err != nil {
			return nil
		}
		return val
	},
	"core::integer::u32": func(typeName string, data string) interface{} {
		val, err := strconv.ParseUint(data, 0, 32)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return val
	},
	"core::integer::u64": func(typeName string, data string) interface{} {
		val, err := strconv.ParseUint(data, 0, 64)
		if err != nil {
			return nil
		}
		return val
	},
	"core::integer::u128": func(typeName string, data string) interface{} {
		// TODO: not u128
		val, err := strconv.ParseUint(data, 0, 64)
		if err != nil {
			return nil
		}
		return val
	},
	"core::integer::u256": func(typeName string, data string) interface{} {
		// TODO
		val, err := strconv.ParseUint(data, 0, 256)
		if err != nil {
			return nil
		}
		return val
	},
	"core::integer::i8": func(typeName string, data string) interface{} {
		val, err := strconv.ParseInt(data, 0, 8)
		if err != nil {
			return nil
		}
		return val
	},
	"core::integer::i16": func(typeName string, data string) interface{} {
		val, err := strconv.ParseInt(data, 0, 16)
		if err != nil {
			return nil
		}
		return val
	},
	"core::integer::i32": func(typeName string, data string) interface{} {
		val, err := strconv.ParseInt(data, 0, 32)
		if err != nil {
			return nil
		}
		return val
	},
	"core::integer::i64": func(typeName string, data string) interface{} {
		val, err := strconv.ParseInt(data, 0, 64)
		if err != nil {
			return nil
		}
		return val
	},
	"core::integer::i128": func(typeName string, data string) interface{} {
		val, err := strconv.ParseInt(data, 0, 128)
		if err != nil {
			return nil
		}
		return val
	},
	"core::bool": func(typeName string, data string) interface{} {
		boolUint, err := strconv.ParseUint(data, 0, 8)
		if err != nil {
			return nil
		}
		if boolUint == 0 {
			return false
		} else {
			return true
		}
	},
	"core::starknet::contract_address::ContractAddress": func(typeName string, data string) interface{} {
		return data
	},
	"core::starknet::class_hash::ClassHash": func(typeName string, data string) interface{} {
		return data
	},
}
