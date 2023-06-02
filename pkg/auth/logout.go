package auth

import (
	"fmt"
	"os"
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
		err = os.WriteFile(polarisHostsFilepath, []byte("{}"), os.ModePerm)
		if err != nil {
			return fmt.Errorf("writing data to file: %w", err)
		}
		return nil
	}
	return nil
}
