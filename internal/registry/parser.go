package registry

import (
	"fmt"
	"strconv"
)

func GetEventTypeName(eventSelector string, abi interface{}) (string, error) {
  if abi == nil {
    return "", fmt.Errorf("abi is nil")
  }
  for _, event := range abi.([]interface{}) {
  }
  return "", fmt.Errorf("event not found")
}

// TODO: Check valid data
// TODO: Parse data based on type
func StarknetTypeDataMin(typeName string, abi interface{}, data []string) interface{} {
	if IsPrimitiveType(typeName) {
		return StarknetStringToTypedData(typeName, data[0])
	} else if IsArrayType(typeName) {
		return data
	} else if IsStructType(typeName) {
		fields := map[string]interface{}{}
		for _, abi := range abi.([]interface{}) {
			if abi.(map[string]interface{})["name"] == typeName ||
				stringEndsWith(abi.(map[string]interface{})["name"].(string), "::"+typeName) {
				for i, member := range abi.(map[string]interface{})["members"].([]interface{}) {
					memberName := member.(map[string]interface{})["name"]
					fields[memberName.(string)] = StarknetTypeDataMin(member.(map[string]interface{})["type"].(string), abi, []string{data[i]})
				}
				break
			}
		}
		return fields
	}
	return nil
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
		if typeName == array.Type || stringStartsWith(typeName, array.Type+"<") {
			return true
		}
	}
	return false
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
	},
}

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
		val, err := strconv.ParseUint(data, 16, 64)
		if err != nil {
			return nil
		}
		return val
	},
	"core::integer::u128": func(typeName string, data string) interface{} {
		val, err := strconv.ParseUint(data, 0, 128)
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
		val, err := strconv.ParseBool(data)
		if err != nil {
			return nil
		}
		return val
	},
	"core::starknet::contract_address::ContractAddress": func(typeName string, data string) interface{} {
		return data
	},
}
