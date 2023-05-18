package auth

import (
	"errors"
	"fmt"
	"net"
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
const registerPath = "/auth/register"

const (
	loginUsingBrowser          = "Login with a web browser"
	pasteAnAuthenticationToken = "Paste an authentication token"
)

type paramsOrError struct {
	token        string
	user         string
	organization string
	err          error
}

type Host struct {
	Token        string `yaml:"token"`
	User         string `yaml:"user"`
	Organization string `yaml:"organization"`
}

var paramsOrErrorChan = make(chan paramsOrError)

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

	var user, token, organization string
	if answer == loginUsingBrowser {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}
		localServerPort := listener.Addr().(*net.TCPAddr).Port
		err = openBrowser(fmt.Sprintf(insightsURL + registerPath + "?source=polaris&callbackUrl=" + fmt.Sprintf("http://localhost:%d/auth/login/callback", localServerPort)))
		if err != nil {
			logrus.Fatal(err)
		}

		var router *mux.Router
		go func() {
			router = mux.NewRouter()
			router.HandleFunc("/auth/login/callback", callbackHandler(localServerPort))
			if err := http.Serve(listener, router); err != nil {
				paramsOrErrorChan <- paramsOrError{err: fmt.Errorf("starting the local http server: %w", err)}
			}
		}()

		// wait the browser to callback the local server
		paramOrError := <-paramsOrErrorChan

		if paramOrError.err != nil {
			return paramOrError.err
		}

		token = paramOrError.token
		user = paramOrError.user
		organization = paramOrError.organization
	} else {
		var answer string
		err := survey.AskOne(&survey.Password{Message: "Paste your authentication token:"}, &answer, survey.WithValidator(validateToken))
		if err != nil {
			return fmt.Errorf("asking how to authenticate: %w", err)
		}
		token = answer
		user = "admin"           // TODO: Vitor - fetch name from bots endpoint
		organization = "acme-co" // TODO: Vitor - fetch organization from bots endpoint
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

	content := map[string]Host{insightsURL: {Token: token, User: user, Organization: organization}}
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
	fmt.Printf("✓ Logged in organization %s as %s.\n", organization, user)
	return nil
}

func callbackHandler(localServerPort int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		var err error
		if len(token) == 0 {
			err = errors.New("token query param is required in callback")
		}
		user := r.URL.Query().Get("user")
		if len(user) == 0 {
			err = errors.New("user query param is required in callback")
		}
		organization := r.URL.Query().Get("organization")
		if len(organization) == 0 {
			err = errors.New("organization query param is required in callback")
		}
		if err != nil {
			fmt.Fprintf(w, "unable to perform integration: %v", err)
			paramsOrErrorChan <- paramsOrError{err: err}
			return
		}
		fmt.Fprint(w, "integration finished successfully, you can safely close this tab now")
		paramsOrErrorChan <- paramsOrError{token: token, user: user, organization: organization}
		return
	}
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
