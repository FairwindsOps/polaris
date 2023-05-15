package cmd

import (
	"github.com/fairwindsops/polaris/pkg/auth"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(tokenCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate polaris with Fairwinds Insights",
	Long:  `Authenticate polaris with Fairwinds Insights so better experience`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate polaris with Fairwinds Insights.",
	Long:  `Authenticate polaris with Fairwinds Insights.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := auth.HandleLogin()
		if err != nil {
			logrus.Fatal(err)
		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of a Fairwinds Insights.",
	Long:  `Log out of a Fairwinds Insights.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := auth.HandleLogout()
		if err != nil {
			logrus.Fatal(err)
		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "View authentication status.",
	Long:  `View authentication status.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := auth.PrintStatus()
		if err != nil {
			logrus.Fatalf("printing status")
		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Print the auth token gh is configured to use.",
	Long:  `Print the auth token gh is configured to use.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := auth.PrintToken()
		if err != nil {
			logrus.Fatalf("printing token: %w", err)
		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
