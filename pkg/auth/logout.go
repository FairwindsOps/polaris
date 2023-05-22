package auth

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func HandleLogout() error {
	if !IsLoggedIn() {
		fmt.Println("not logged in to Fairwinds Insights")
		return nil
	}
	err := performLogout()
	if err != nil {
		return fmt.Errorf("performing logout: %v", err)
	}
	fmt.Println("✓ Logged out of Fairwinds Insights")
	return nil
}

func performLogout() error {
	content, err := readPolarisHostsFile()
	if err != nil {
		return nil
	}

	if len(content) > 0 {
		f, err := os.Create(polarisHostsFilepath)
		if err != nil {
			return nil
		}
		defer func() {
			if err := f.Close(); err != nil {
				logrus.Fatalf("closing user polaris hosts file: %v", err)
			}
		}()

		_, err = f.Write([]byte("{}"))
		if err != nil {
			return fmt.Errorf("writing data to file: %w", err)
		}
		return nil
	}
	return nil
}
