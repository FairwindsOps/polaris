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
func DeployHandler(w http.ResponseWriter, r *http.Request, c conf.Configuration) {
	var results []Results
	clientset := kube.CreateClientset()
	deploys, err := clientset.AppsV1().Deployments("").List(metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	for _, deploy := range deploys.Items {
		results = append(results, ValidateDeploys(c, &deploy))
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
