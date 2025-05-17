package routes

import (
	"net/http"

	"github.com/b-j-roberts/foc-engine/internal/provider"
	"github.com/b-j-roberts/foc-engine/internal/registry"
	routeutils "github.com/b-j-roberts/foc-engine/routes/utils"
)

func InitRegistryRoutes() {
  http.HandleFunc("/registry/add-registry-contract", AddRegistryContract)
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
