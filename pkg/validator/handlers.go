package validator

import (
	"encoding/json"
	"fmt"
	"net/http"

	conf "github.com/reactiveops/fairwinds/pkg/config"
)

// PingHandler is an endpoint to check the server is up.
func PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong\n")
}

// DeployHandler creates a handler for to validate the current deploy workloads.
func DeployHandler(w http.ResponseWriter, r *http.Request, c conf.Configuration) {
	results, _ := ValidateDeploys(c)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
