package auth

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const loginPath = "/auth/login"
const authServerPort = "8089"

var userHomeDir string

type tokenOrError struct {
	token string
	err   error
}

type Host struct {
	Token string `yaml:"token"`
}

var tokenOrErrorChan = make(chan tokenOrError)

func init() {
	var err error
	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("reading user home dir: %v", err)
	}
}

func HandleLogin() error {
	// checking for existing configuration

	polarisHostsFilePath := userHomeDir + "/.config/polaris/hosts.yaml"
	if _, err := os.Stat(polarisHostsFilePath); err == nil {
		// file exists
		f, err := os.Open(polarisHostsFilePath)
		if err != nil {
			return fmt.Errorf("opening existing polaris hosts file: %w", err)
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("reading from existing polaris hosts file: %w", err)
		}

		content := map[string]Host{}
		err = yaml.Unmarshal(b, &content)
		if err != nil {
			return fmt.Errorf("unmarshaling from existing polaris hosts file: %w", err)
		}

		if len(content) > 0 {
			var reAuthenticate bool
			prompt := &survey.Confirm{Message: fmt.Sprintf("You're already logged into %s. Do you want to re-authenticate?", insightsURL)}
			err = survey.AskOne(prompt, &reAuthenticate)
			if err != nil {
				return fmt.Errorf("prompting re-authenticate: %w", err)
			}
			if !reAuthenticate {
				// bail-out
				return nil
			}
		}
	}

	err := openBrowser(fmt.Sprintf(insightsURL + loginPath + "?source=polaris"))
	if err != nil {
		logrus.Fatal(err)
	}

	var router *mux.Router
	go func() {
		router = mux.NewRouter()
		router.HandleFunc("/auth/login/callback", callbackHandler)
		if err := http.ListenAndServe(":"+authServerPort, router); err != nil {
			tokenOrErrorChan <- tokenOrError{err: fmt.Errorf("starting the local http server: %w", err)}
		}
	}()

	// wait the browser to callback the local server
	tokenOrError := <-tokenOrErrorChan

	if tokenOrError.err != nil {
		return tokenOrError.err
	}
	logrus.Debugf("got token: %s", tokenOrError.token)

	polarisCfgDir := filepath.Join(userHomeDir, ".config", "polaris")
	err = os.MkdirAll(polarisCfgDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("creating polaris config dir: %w", err)
	}

	f, err := os.Create(filepath.Join(polarisCfgDir, "hosts.yaml"))
	if err != nil {
		return fmt.Errorf("opening user polaris hosts file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.Fatalf("closing user polaris hosts file: %v", err)
		}
	}()

	content := map[string]Host{insightsURL: {Token: tokenOrError.token}}
	b, err := yaml.Marshal(content)
	if err != nil {
		return fmt.Errorf("marshalling yaml data: %w", err)
	}

	_, err = f.Write(b)
	if err != nil {
		return fmt.Errorf("writing data to file: %w", err)
	}

	logrus.Debugf("hosts file has been saved")
	return nil
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		tokenOrErrorChan <- tokenOrError{err: errors.New("token query param is required in callback")}
		return
	}
	tokenOrErrorChan <- tokenOrError{token: token}
	return
}
