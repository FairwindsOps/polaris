package auth

import "fmt"

func PrintStatus() error {

	if content, err := readPolarisHostsFile(); err == nil {
		if len(content) > 0 {
			for k, h := range content {
				fmt.Println(k, h)
				// TODO: Vitor - implement this
				/*
					✓ Logged in to github.com as vitorvezani (keyring)
					✓ Git operations for github.com configured to use ssh protocol.
					✓ Token: ghp_************************************
					✓ Token scopes: admin:public_key, read:org, repo
				*/
			}
			return nil
		}
	}
	fmt.Println("You are not logged into Fairwinds Insights. Run polaris auth login to authenticate.")
	return nil
}
