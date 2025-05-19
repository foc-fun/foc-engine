package routes

import (
	"encoding/json"
	"net/http"

	"github.com/b-j-roberts/foc-engine/internal/provider"
	"github.com/b-j-roberts/foc-engine/internal/registry"
	routeutils "github.com/b-j-roberts/foc-engine/routes/utils"
)

func InitRegistryRoutes() {
	http.HandleFunc("/registry/add-registry-contract", AddRegistryContract)
	http.HandleFunc("/registry/get-registry-contracts", GetRegistryContracts)
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

	err = provider.SubscribeEvents(registryContractAddress)
	if err != nil {
		routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to subscribe to events")
		return
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
