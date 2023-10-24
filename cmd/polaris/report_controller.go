package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

func init() {
	rootCmd.AddCommand(reportControllerCommand)
	utilruntime.Must(clientgoscheme.AddToScheme(clientgoscheme.Scheme))
}

var reportControllerCommand = &cobra.Command{
	Use:   "report-controller",
	Short: "Runs the Insights report controller.",
	Long:  "Runs the Insights report controller.",
	Run: func(cmd *cobra.Command, args []string) {

		mgr, err := ctrl.NewManager(k8sConfig.GetConfigOrDie(), ctrl.Options{
			Scheme:                 clientgoscheme.Scheme,
			MetricsBindAddress:     ":8080",
			Port:                   9443,
			HealthProbeBindAddress: ":8081",
			LeaderElection:         false,
			LeaderElectionID:       "fc2613a0.fairwinds.com",
			Namespace:              "default",
		})
		if err != nil {
			logrus.Errorf("Unable to set up overall controller manager: %v", err)
			os.Exit(1)
		}

		logrus.Infof("Polaris report controller running")

		if err = (&DeploymentReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			logrus.Error(err, "unable to create controller", "controller", "Deployment")
			os.Exit(1)
		}

		if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			logrus.Error(err, "unable to set up health check")
			os.Exit(1)
		}
		if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			logrus.Error(err, "unable to set up ready check")
			os.Exit(1)
		}

		logrus.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			logrus.Error(err, "problem running manager")
			os.Exit(1)
		}
	},
}
