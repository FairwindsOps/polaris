package auth

import (
	"os"

	"github.com/sirupsen/logrus"
)

const insightsURL = "http://localhost:3001" // TODO: make it configurable

var userHomeDir string
var polarisHostsFilepath string

func init() {
	var err error
	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("reading user home dir: %v", err)
	}
	polarisHostsFilepath = userHomeDir + "/.config/polaris/hosts.yaml"
}
