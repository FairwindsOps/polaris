package auth

import (
	"fmt"
	"strings"
)

func PrintStatus() error {
	if content, err := readPolarisHostsFile(); err == nil {
		if len(content) > 0 {
			for k, h := range content {
				if len(content) == 1 {
					fmt.Printf("✓ Logged in to %s as <user>\n", k)
					fmt.Printf("✓ Token: %s\n", hideToken(h.Token, 3))
				}
			}
			return nil
		}
	}
	fmt.Println("You are not logged into Fairwinds Insights. Run polaris auth login to authenticate.")
	return nil
}

func hideToken(token string, hideAfter int) string {
	var i int
	return strings.Map(func(r rune) rune {
		defer func() {
			i++
		}()
		if i > hideAfter {
			return []rune("*")[0]
		}
		return r
	}, token)
}
