package auth

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func HandleLogout() error {
	ok, err := performLogout()
	if err != nil {
		return fmt.Errorf("performing logout: %v", err)
	}
	if ok {
		fmt.Println("✓ Logged out of Fairwinds Insights")
	} else {
		fmt.Println("not logged in to Fairwinds Insights")
	}
	return nil
}

func performLogout() (bool, error) {
	if _, err := os.Stat(polarisHostsFilepath); err == nil {
		f, err := os.Open(polarisHostsFilepath)
		if err != nil {
			return false, nil
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return false, nil
		}

		content := map[string]Host{}
		err = yaml.Unmarshal(b, &content)
		if err != nil {
			return false, nil
		}

		if len(content) > 0 {
			f, err := os.Create(polarisHostsFilepath)
			if err != nil {
				return false, nil
			}
			defer func() {
				if err := f.Close(); err != nil {
					logrus.Fatalf("closing user polaris hosts file: %v", err)
				}
			}()

			_, err = f.Write([]byte("{}"))
			if err != nil {
				return false, fmt.Errorf("writing data to file: %w", err)
			}
			return true, nil
		}
	}
	return false, nil
}
