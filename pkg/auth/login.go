package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fairwindsops/polaris/pkg/insights"
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

func HandleLogin(insightsHost string) error {
	if _, err := os.Stat(polarisHostsFilepath); err == nil {
		content, err := readPolarisHostsFile()
		if err != nil {
			return fmt.Errorf("reading polaris hosts file: %w", err)
		}

		if len(content) > 0 {
			if h, ok := content[insightsHost]; ok {
				c := insights.NewHTTPClient(insightsHost, h.Organization, h.Token)
				isValid, err := c.IsTokenValid()
				if err != nil {
					return err
				}
				if isValid {
					var reAuthenticate bool
					err = survey.AskOne(&survey.Confirm{Message: fmt.Sprintf("You're already logged into %s. Do you want to re-authenticate?", insightsHost)}, &reAuthenticate)
					if err != nil {
						return fmt.Errorf("prompting re-authenticate: %w", err)
					}
					if !reAuthenticate {
						// bail-out
						return nil
					}
				}
			}
		}
	}

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
		url := fmt.Sprintf(insightsHost + registerPath + "?source=polaris&callbackUrl=" + fmt.Sprintf("http://localhost:%d/auth/login/callback", localServerPort))
		err = openBrowser(url)
		if err != nil {
			logrus.Warnf("could not open browser: %v", err)
			logrus.Infoln("paste the link below into your browser:")
			os.Stdout.Write([]byte(url + "\n"))
		}

		var router *mux.Router
		go func() {
			router = mux.NewRouter()
			router.HandleFunc("/auth/login/callback", callbackHandler(insightsHost, localServerPort))
			if err := http.Serve(listener, router); err != nil {
				paramsOrErrorChan <- paramsOrError{err: fmt.Errorf("starting the local http server: %w", err)}
			}
		}()

		// wait the browser to callback the local server
		paramOrError := <-paramsOrErrorChan

		if paramOrError.err != nil {
			return paramOrError.err
		}

		user = paramOrError.user
		organization = paramOrError.organization
		token = paramOrError.token
	} else {
		var answer string
		var bot bot
		err := survey.AskOne(&survey.Password{Message: "Paste your authentication token:"}, &answer, survey.WithValidator(validateToken(insightsHost, &bot)))
		if err != nil {
			return fmt.Errorf("asking how to authenticate: %w", err)
		}
		token = answer
		user = bot.Name
		organization = bot.Organization
	}

	polarisCfgDir := filepath.Join(userHomeDir, ".config", "polaris")
	err = os.MkdirAll(polarisCfgDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("creating polaris config dir: %w", err)
	}

	content := map[string]Host{insightsHost: {Token: token, User: user, Organization: organization}}
	b, err := yaml.Marshal(content)
	if err != nil {
		return fmt.Errorf("marshalling yaml data: %w", err)
	}

	err = os.WriteFile(polarisHostsFilepath, b, os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing data to file: %w", err)
	}

	logrus.Debugf("hosts file has been saved")

	fmt.Println("✓ Authentication complete.")
	fmt.Printf("✓ Logged in organization %s as %s.\n", organization, user)
	return nil
}

func fetchAuthToken(insightsHost, organization, code string) (string, error) {
	authTokenURL := fmt.Sprintf("%s/v0/organizations/%s/auth/token", insightsHost, organization)
	body := map[string]any{"grantType": "authorization_code", "code": code}
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	r, err := http.NewRequest("POST", authTokenURL, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	r.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return "", fmt.Errorf("expected 200 OK - received %s", res.Status)
	}

	var rBody map[string]any
	err = json.NewDecoder(res.Body).Decode(&rBody)
	if err != nil {
		return "", err
	}

	token, ok := rBody["accessToken"].(string)
	if !ok {
		return "", fmt.Errorf("unable to parse accessToken from response body: %v", rBody)
	}

	return token, nil
}

func callbackHandler(insightsHost string, localServerPort int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// checks for error in the params
		errMsg := r.URL.Query().Get("error")
		if len(errMsg) > 0 {
			errDescriptionMsg := r.URL.Query().Get("error_description")
			fmt.Fprintf(w, "unable to perform integration: %s - %s", errMsg, errDescriptionMsg)
			paramsOrErrorChan <- paramsOrError{err: fmt.Errorf("%s - %s", errMsg, errDescriptionMsg)}
			return
		}

		var err error
		code := r.URL.Query().Get("code")
		if len(code) == 0 {
			err = errors.New("code query param is required in callback")
		}
		user := r.URL.Query().Get("user")
		if len(user) == 0 {
			err = errors.New("user query param is required in callback")
		}
		organization := r.URL.Query().Get("organization")
		if len(organization) == 0 {
			err = errors.New("organization query param is required in callback")
		}
		token, err := fetchAuthToken(insightsHost, organization, code)
		if err != nil {
			err = fmt.Errorf("fetching auth token: %w", err)
		}
		if err != nil {
			fmt.Fprintf(w, "unable to perform integration: %v", err)
			paramsOrErrorChan <- paramsOrError{err: err}
			return
		}

		fmt.Fprint(w, "Polaris and Fairwinds Insights integration has finished successfully, your credentials are set! You can safely close this tab now.")
		paramsOrErrorChan <- paramsOrError{token: token, user: user, organization: organization}
		return
	}
}

func validateToken(insightsHost string, bot *bot) func(args any) error {
	return func(args any) error {
		token, ok := args.(string)
		if !ok {
			return errors.New("casting token to string")
		}
		if len(strings.TrimSpace(token)) <= 0 {
			return errors.New("token is required")
		}
		return fetchOrganizationBot(insightsHost, token, bot)
	}
}

type bot struct {
	ID           int
	Organization string
	Name         string
	Role         string
	AuthToken    string
	CreatedAt    time.Time
}

func fetchOrganizationBot(insightsHost, authToken string, bot *bot) error {
	authTokenURL := fmt.Sprintf("%s/v0/bots/from-request", insightsHost)
	r, err := http.NewRequest("GET", authTokenURL, nil)
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", "Bearer "+authToken)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return fmt.Errorf("expected 200 (OK) - received %d", res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(bot)
	if err != nil {
		return err
	}
	return nil
}

func IsLoggedIn() bool {
	if _, err := os.Stat(polarisHostsFilepath); err == nil {
		content, err := readPolarisHostsFile()
		if err != nil {
			return false
		}
		return len(content) > 0
	}
	return false
}
