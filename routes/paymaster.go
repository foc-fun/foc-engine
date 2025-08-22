package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/NethermindEth/starknet.go/typedData"
	routeutils "github.com/b-j-roberts/foc-engine/routes/utils"
)

func InitPaymasterRoutes() {
	http.HandleFunc("/paymaster/build-gasless-tx", BuildGaslessTx)
	http.HandleFunc("/paymaster/send-gasless-tx", SendGaslessTx)
}

type Call struct {
	ContractAddress string   `json:"contractAddress"`
	Entrypoint      string   `json:"entrypoint"`
	Calldata        []string `json:"calldata"`
}

type DeploymentData struct {
	ClassHash string   `json:"class_hash"`
	Calldata  []string `json:"calldata"`
	Salt      string   `json:"salt"`
	Unique    string   `json:"unique"`
}

type GaslessTxInput struct {
	Account        string          `json:"account"`
	Calls          []Call          `json:"calls"`
	Network        string          `json:"network"`
	DeploymentData *DeploymentData `json:"deploymentData"`
}

func BuildGaslessTx(w http.ResponseWriter, r *http.Request) {
	input, err := routeutils.ReadJsonBody[GaslessTxInput](r)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// TODO: Check api key is set
	// TODO: Perform call checking and validation
	// TODO: Move this to a separate package for better organization
	baseUrl := "https://sepolia.api.avnu.fi" // https://starknet.api.avnu.fi
	avnuPaymasterUrl := baseUrl + "/paymaster/v1/build-typed-data"
	requestBody := map[string]interface{}{
		"userAddress": input.Account,
		"calls":       input.Calls, // TODO: Format calls if necessary ( compile & hexify )
	}
	requestHeaders := map[string]string{
		"api-key": os.Getenv("AVNU_PAYMASTER_API_KEY"), // TODO: Allow usage of different API keys for diff inputs?
	}

	if input.DeploymentData != nil {
		requestBody["accountClassHash"] = input.DeploymentData.ClassHash
	}
	res, err := routeutils.PostJson(avnuPaymasterUrl, requestBody, requestHeaders)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to build gasless transaction")
		return
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to read response body")
		return
	}
	defer res.Body.Close()
	txJsonBytes := buf.Bytes()

	// Parse the typed data to get the message hash
	var td typedData.TypedData
	err = json.Unmarshal(txJsonBytes, &td)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to parse typed data")
		return
	}

	// Get the message hash using the account address
	messageHash, err := td.GetMessageHash(input.Account)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to calculate message hash")
		return
	}

	// Return both the typed data and the message hash
	response := map[string]interface{}{
		"typedData":   json.RawMessage(txJsonBytes),
		"messageHash": messageHash.String(),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to marshal response")
		return
	}

	routeutils.WriteDataJson(w, string(responseBytes))
}

type SendGaslessTxInput struct {
	Account        string          `json:"account"`
	TxData         string          `json:"txData"`    // Value returned from BuildGaslessTx ( TODO: Cache? )
	Signature      []string        `json:"signature"` // Users signature of the transaction data
	Network        string          `json:"network"`
	DeploymentData *DeploymentData `json:"deploymentData"`
}

func SendGaslessTx(w http.ResponseWriter, r *http.Request) {
	input, err := routeutils.ReadJsonBody[SendGaslessTxInput](r)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	baseUrl := "https://sepolia.api.avnu.fi" // https://starknet.api.avnu.fi
	avnuPaymasterUrl := baseUrl + "/paymaster/v1/execute"
	requestBody := map[string]interface{}{
		"userAddress": input.Account,
		"typedData":   input.TxData,
		"signature":   input.Signature,
	}
	requestHeaders := map[string]string{
		"api-key": os.Getenv("AVNU_PAYMASTER_API_KEY"), // TODO: Allow usage of different API keys for diff inputs?
	}
	if input.DeploymentData != nil {
		requestBody["deploymentData"] = map[string]interface{}{
			"class_hash": input.DeploymentData.ClassHash,
			"calldata":   input.DeploymentData.Calldata,
			"salt":       input.DeploymentData.Salt,
			"unique":     input.DeploymentData.Unique,
		}
	}
	res, err := routeutils.PostJson(avnuPaymasterUrl, requestBody, requestHeaders)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to send gasless transaction")
		return
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to read response body")
		return
	}
	defer res.Body.Close()

	responseBytes := buf.Bytes()

	routeutils.WriteDataJson(w, string(responseBytes))
}
