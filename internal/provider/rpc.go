package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/b-j-roberts/foc-engine/internal/config"
)

type StarknetRpcCall struct {
  ID      int         `json:"id"`
  Jsonrpc string      `json:"jsonrpc"`
  Method  string      `json:"method"`
  Params  interface{} `json:"params"`
}

type StarknetRpcResponse struct {
  Jsonrpc string      `json:"jsonrpc"`
  ID      int         `json:"id"`
  Result  interface{} `json:"result"`
  Error   interface{} `json:"error"`
}

/*
  Sample curl command to replicate

  curl --location 'http://localhost:5050' \               ✔
       --header 'accept: application/json' \
       --header 'content-type: application/json' \
       --data '{
           "id": 1,
           "jsonrpc": "2.0",
           "method": "starknet_blockNumber",
           "params": []
       }'
       {"jsonrpc":"2.0","id":1,"result":1}
*/
func GetStarknetLatestBlockNumber() (uint64, error) {
  // Create a new StarknetRpcCall object
  call := StarknetRpcCall{
    ID:      1,
    Jsonrpc: "2.0",
    Method:  "starknet_blockNumber",
    Params:  []interface{}{},
  }

  // Marshal the call to JSON
  jsonData, err := json.Marshal(call)
  if err != nil {
    return 0, err
  }

  // Send the request to the Starknet RPC endpoint
  url := "http://" + config.Conf.Rpc.Host
  resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
  if err != nil {
    return 0, err
  }
  defer resp.Body.Close()

  // Read the response body
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return 0, err
  }

  // Unmarshal the response into a StarknetRpcResponse object
  var response StarknetRpcResponse
  if err := json.Unmarshal(body, &response); err != nil {
    return 0, err
  }

  // Check for errors in the response
  if response.Error != nil {
    return 0, fmt.Errorf("error from server: %v", response.Error)
  }

  // Extract the block number from the result
  blockNumber, ok := response.Result.(float64)
  if !ok {
    return 0, fmt.Errorf("invalid result format")
  }

  return uint64(blockNumber), nil
}
