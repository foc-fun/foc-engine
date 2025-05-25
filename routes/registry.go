package routes

import (
	"encoding/json"
	"net/http"

	"github.com/b-j-roberts/foc-engine/internal/db/mongo"
	"github.com/b-j-roberts/foc-engine/internal/provider"
	"github.com/b-j-roberts/foc-engine/internal/registry"
	routeutils "github.com/b-j-roberts/foc-engine/routes/utils"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitRegistryRoutes() {
	http.HandleFunc("/registry/add-registry-contract", AddRegistryContract)
	http.HandleFunc("/registry/get-registry-contracts", GetRegistryContracts)

  http.HandleFunc("/registry/get-registered-contract", GetRegisteredContract)
}

func AddRegistryContract(w http.ResponseWriter, r *http.Request) {
	if routeutils.AdminMiddleware(w, r) {
		routeutils.WriteErrorJson(w, http.StatusUnauthorized, "Only the admin can add registry contracts")
		return
	}

	jsonBody, err := routeutils.ReadJsonBody[map[string]string](r)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	registryContractAddress, ok := (*jsonBody)["address"]
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

  if subscribeEvents == "true" {
    err = provider.SubscribeEvents(registryContractAddress)
    if err != nil {
      routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to subscribe to events")
      return
    }
  }
	registry.AddRegistryAddress(registryContractAddress)

	routeutils.WriteResultJson(w, "Registry contract added successfully")
}

func GetRegistryContracts(w http.ResponseWriter, r *http.Request) {
	// RegistryAddresses  map[string]bool
	registeredContractsMap := registry.FocRegistry.RegistryAddresses
	registeredContracts := make([]string, 0, len(registeredContractsMap))
	for address := range registeredContractsMap {
		registeredContracts = append(registeredContracts, address)
	}
	resultJson := map[string]interface{}{
		"registry_contracts": registeredContracts,
	}
	resultJsonBytes, err := json.Marshal(resultJson)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to marshal JSON")
		return
	}

	routeutils.WriteDataJson(w, string(resultJsonBytes))
}

func GetRegisteredContract(w http.ResponseWriter, r *http.Request) {
  contractName := r.URL.Query().Get("contractName")
  if contractName == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing 'contractName' query parameter")
    return
  }
  contractVersion := r.URL.Query().Get("contractVersion")
  if contractVersion == "" {
    contractVersion = "latest" // Default to latest if not provided
  }
  findOptions := options.Find().SetSort(map[string]interface{}{
    "_id": -1,
  }).SetLimit(1)
  // TODO: Have registry setup like events w/ typed data
  res, err := mongo.GetFocEngineRegistryCollection().Find(r.Context(), map[string]interface{}{
    "contract.name": contractName,
    "contract.version": contractVersion,
  }, findOptions)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to query database")
    return
  }
  defer res.Close(r.Context())

  var event map[string]interface{}
  if res.Next(r.Context()) {
    if err := res.Decode(&event); err != nil {
      routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to decode database response")
      return
    }
  } else {
    routeutils.WriteErrorJson(w, http.StatusNotFound, "Contract not found")
    return
  }
  if err := res.Err(); err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to iterate over database response")
    return
  }
  resultJsonBytes, err := json.Marshal(event)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to marshal JSON")
    return
  }
  routeutils.WriteDataJson(w, string(resultJsonBytes))
}
