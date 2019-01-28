package validator

import (
	"encoding/json"
	"fmt"
	"net/http"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PingHandler is an endpoint to check the server is up.
func PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong\n")
}

// DeployHandler creates a handler for to validate the current deploy workloads.
func DeployHandler(w http.ResponseWriter, r *http.Request, c conf.Configuration) error {
	var results []Results
	deploys, err := kube.AppsV1API.Deployments("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deploy := range deploys.Items {
		result := ValidateDeploys(c, &deploy, Results{})
		results = append(results, result)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
	return nil
}
