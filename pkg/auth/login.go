package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const loginPath = "/auth/login"
const authServerPort = "8089"

const (
	loginUsingBrowser          = "Login with a web browser"
	pasteAnAuthenticationToken = "Paste an authentication token"
)

type tokenOrError struct {
	token string
	err   error
}

type Host struct {
	Token string `yaml:"token"`
}

var tokenOrErrorChan = make(chan tokenOrError)

func HandleLogin() error {
	if _, err := os.Stat(polarisHostsFilepath); err == nil {
		content, err := readPolarisHostsFile()
		if err != nil {
			return fmt.Errorf("reading polaris hosts file: %w", err)
		}

		if len(content) > 0 {
			var reAuthenticate bool
			err = survey.AskOne(&survey.Confirm{Message: fmt.Sprintf("You're already logged into %s. Do you want to re-authenticate?", insightsURL)}, &reAuthenticate)
			if err != nil {
				return fmt.Errorf("prompting re-authenticate: %w", err)
			}
			if !reAuthenticate {
				// bail-out
				return nil
			}
		}
	}

	// How would you like to authenticate GitHub CLI?  [Use arrows to move, type to filter]
	selection := &survey.Select{
		Message: "How would you like to authenticate Polaris?",
		Options: []string{loginUsingBrowser, pasteAnAuthenticationToken},
		Default: loginUsingBrowser,
	}

	var answer string
	err := survey.AskOne(selection, &answer)
	if err != nil {
		return fmt.Errorf("asking how to authenticate: %w", err)
	}

	var token string
	if answer == loginUsingBrowser {
		err = openBrowser(fmt.Sprintf(insightsURL + loginPath + "?source=polaris"))
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

		token = tokenOrError.token
	} else {
		var answer string
		err := survey.AskOne(&survey.Password{Message: "Paste your authentication token:"}, &answer, survey.WithValidator(validateToken))
		if err != nil {
			return fmt.Errorf("asking how to authenticate: %w", err)
		}
		token = answer
	}

	polarisCfgDir := filepath.Join(userHomeDir, ".config", "polaris")
	err = os.MkdirAll(polarisCfgDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("creating polaris config dir: %w", err)
	}

	f, err := os.Create(polarisHostsFilepath)
	if err != nil {
		return fmt.Errorf("opening user polaris hosts file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.Fatalf("closing user polaris hosts file: %v", err)
		}
	}()

	content := map[string]Host{insightsURL: {Token: token}}
	b, err := yaml.Marshal(content)
	if err != nil {
		return fmt.Errorf("marshalling yaml data: %w", err)
	}

	_, err = f.Write(b)
	if err != nil {
		return fmt.Errorf("writing data to file: %w", err)
	}

	logrus.Debugf("hosts file has been saved")

	fmt.Println("✓ Authentication complete.")
	fmt.Println("✓ Logged in as <user>.")
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

func validateToken(args any) error {
	token, ok := args.(string)
	if !ok {
		return errors.New("casting token to string")
	}

	if len(strings.TrimSpace(token)) <= 0 {
		return errors.New("token is required")
	}

	// TODO: Vitor - validate token against the API?
	return nil
}
