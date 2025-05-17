package routes

import (
	"fmt"
	"net/http"

	routeutils "github.com/b-j-roberts/foc-engine/routes/utils"
)

func InitBaseRoutes() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    routeutils.SetupHeaders(w)
    w.WriteHeader(http.StatusOK)
  })
}

func InitRoutes() {
  InitBaseRoutes()
  InitRegistryRoutes()
}

func StartServer(host string, port int) {
  InitRoutes()
  addr := fmt.Sprintf("%s:%d", host, port)
  fmt.Printf("Starting server on %s\n", addr)
  go func() {
    if err := http.ListenAndServe(addr, nil); err != nil {
      fmt.Println("Error starting server:", err)
    }
    fmt.Println("Server stopped")
  }()
}
