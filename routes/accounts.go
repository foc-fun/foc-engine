package routes

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/b-j-roberts/foc-engine/internal/accounts"
	"github.com/b-j-roberts/foc-engine/internal/db/mongo"
	"github.com/b-j-roberts/foc-engine/internal/provider"
	"github.com/b-j-roberts/foc-engine/internal/registry"
	routeutils "github.com/b-j-roberts/foc-engine/routes/utils"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitAccountsRoutes() {
	http.HandleFunc("/accounts/add-accounts-contract", AddAccountsContract)
	http.HandleFunc("/accounts/get-accounts-contracts", GetAccountsContracts)

	http.HandleFunc("/accounts/get-account", GetFocAccount)
	http.HandleFunc("/accounts/mint-funds", MintFunds)
}

func readFeltString(data string) (string, error) {
	decodedData, err := hex.DecodeString(data[2:])
	if err != nil {
		return "", err
	}
	trimmedName := []byte{}
	trimming := true
	for _, b := range decodedData {
		if b == 0 && trimming {
			continue
		}
		trimming = false
		trimmedName = append(trimmedName, b)
	}
	feltString := string(trimmedName)
	return feltString, nil
}

func AddAccountsContract(w http.ResponseWriter, r *http.Request) {
	if routeutils.AdminMiddleware(w, r) {
		routeutils.WriteErrorJson(w, http.StatusUnauthorized, "Only the admin can add accounts contracts")
		return
	}
	jsonBody, err := routeutils.ReadJsonBody[map[string]string](r)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	accountsContractAddress, ok := (*jsonBody)["address"]
	if !ok {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing 'address' field in JSON body")
		return
	}

	subscribeEvents, ok := (*jsonBody)["subscribeEvents"]
	if !ok {
		subscribeEvents = "false" // Default to false if not provided
	}
	if subscribeEvents != "true" && subscribeEvents != "false" {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid 'subscribeEvents' field in JSON body, must be 'true' or 'false'")
		return
	}

	// Remove leading 0x & all 0s after 0x
	if len(accountsContractAddress) > 2 && accountsContractAddress[:2] == "0x" {
		accountsContractAddress = accountsContractAddress[2:]
	}
	for len(accountsContractAddress) > 0 && accountsContractAddress[0] == '0' {
		accountsContractAddress = accountsContractAddress[1:]
	}
	accountsContractAddress = "0x" + accountsContractAddress

	if subscribeEvents == "true" {
		accountsClassHash, ok := (*jsonBody)["class_hash"]
		if !ok {
			routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing 'class_hash' field in JSON body")
			return
		}

		err = provider.SubscribeEvents(accountsContractAddress)
		if err != nil {
			routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to subscribe to events")
			return
		}
		registry.RegisterContract(accountsContractAddress, accountsClassHash)
	}
	accounts.AddAccountsContract(accountsContractAddress)

	routeutils.WriteResultJson(w, "Accounts contract added successfully")
}

func GetAccountsContracts(w http.ResponseWriter, r *http.Request) {
	// RegistryAddresses  map[string]bool
	accountsContract := accounts.GetAccountsContract()
	resultJson := map[string]interface{}{
		"accounts_contract": accountsContract,
	}
	resultJsonBytes, err := json.Marshal(resultJson)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to marshal JSON")
		return
	}

	routeutils.WriteDataJson(w, string(resultJsonBytes))
}

func GetFocAccount(w http.ResponseWriter, r *http.Request) {
	contractAddress := r.URL.Query().Get("contractAddress")
	if contractAddress == "" {
		contractAddress = accounts.FocAccounts.AccountsContractAddress
	}
	accountAddress := r.URL.Query().Get("accountAddress")
	if accountAddress == "" {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing 'accountAddress' query parameter")
		return
	}
	findOptions := options.Find().SetSort(map[string]interface{}{
		"_id": -1,
	}).SetLimit(1)
	// TODO: Changed username
	// TODO: Cache results?
	res, err := mongo.GetFocEngineEventsCollection().Find(r.Context(), map[string]interface{}{
		"contract_address": accounts.FocAccounts.AccountsContractAddress,
		"contract":         contractAddress,
		"user":             accountAddress,
		"event_type":       "onchain::accounts::FocAccounts::UsernameClaimed",
	}, findOptions)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to query database")
		return
	}
	defer res.Close(r.Context())

	var accountEvent map[string]interface{}
	if res.Next(r.Context()) {
		if err := res.Decode(&accountEvent); err != nil {
			routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to decode event")
			return
		}
	} else {
		routeutils.WriteErrorJson(w, http.StatusNotFound, "Account not found")
		return
	}
	if err := res.Err(); err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to iterate over results")
		return
	}
	username, ok := accountEvent["username"].(string)
	if !ok {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to get username from event")
		return
	}
	// Hex to string utf8
	usernameStr, err := readFeltString(username)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to decode username")
		return
	}
	account := accounts.AccountInfo{
		Username: usernameStr,
	}

	/*
	  account := accounts.GetContractAccountInfo(contractAddress, accountAddress)
	  if account == nil {
	    routeutils.WriteErrorJson(w, http.StatusNotFound, "Account not found")
	    return
	  }
	*/
	resultJson := map[string]interface{}{
		"contract_address": contractAddress,
		"account_address":  accountAddress,
		"account":          account,
	}
	resultJsonBytes, err := json.Marshal(resultJson)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to marshal JSON")
		return
	}
	routeutils.WriteDataJson(w, string(resultJsonBytes))
}

func MintFunds(w http.ResponseWriter, r *http.Request) {
	// curl -X POST http://127.0.0.1:5050/mint -d '{"address":"0x2a15f812c97fbca1bd061ad074ee198877721aba36e09548e43b53b90c77c74","amount":50000000000000000000,"unit":"FRI"}' -H "Content-Type:application/json"
	if routeutils.AdminMiddleware(w, r) {
		routeutils.WriteErrorJson(w, http.StatusUnauthorized, "Only the admin can mint funds")
		return
	}

	jsonBody, err := routeutils.ReadJsonBody[map[string]interface{}](r)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	address, ok := (*jsonBody)["address"].(string)
	if !ok || address == "" {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid 'address' field in JSON body")
		return
	}
	amountStr, ok := (*jsonBody)["amount"].(string)
	if !ok || amountStr == "" {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing or invalid 'amount' field in JSON body")
		return
	}
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok || amount.Cmp(big.NewInt(0)) <= 0 {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid 'amount' field in JSON body")
		return
	}

	unit, ok := (*jsonBody)["unit"].(string)
	if !ok || unit == "" {
		unit = "FRI" // Default to FRI if not provided
	}

	// TODO: Check valid inputs
	provider.Mint(address, amount, unit)

	routeutils.WriteResultJson(w, "Funds minted successfully")
}
