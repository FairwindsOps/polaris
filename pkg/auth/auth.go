package auth

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
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

func readPolarisHostsFile() (map[string]Host, error) {
	f, err := os.Open(polarisHostsFilepath)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	content := map[string]Host{}
	err = yaml.Unmarshal(b, &content)
	return content, err
}
