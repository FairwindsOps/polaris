package auth

import (
	"errors"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var userHomeDir string
var polarisHostsFilepath string

var ErrNotLoggedIn = errors.New("not logged in")

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

func GetAuth(insightsHost string) (*Host, error) {
	hosts, err := readPolarisHostsFile()
	if err != nil {
		return nil, err
	}
	if len(hosts) == 0 {
		return nil, ErrNotLoggedIn
	}
	if h, ok := hosts[insightsHost]; ok {
		return &h, nil
	}
	return nil, ErrNotLoggedIn
}
