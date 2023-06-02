package auth

import (
	"fmt"
	"strings"

	"github.com/fairwindsops/polaris/pkg/insights"
)

func PrintStatus(insightsHost string) error {
	if content, err := readPolarisHostsFile(); err == nil {
		if len(content) > 0 {
			if h, ok := content[insightsHost]; ok {
				c := insights.NewHTTPClient(insightsHost, h.Organization, h.Token)
				isValid, err := c.IsTokenValid()
				if err != nil {
					return err
				}
				if !isValid {
					fmt.Println("✕ Your token is no longer valid. Run polaris auth login to authenticate.")
					return nil
				}
				fmt.Printf("✓ Logged in to %s as %s\n", insightsHost, h.User)
				fmt.Printf("✓ Token: %s\n", hideToken(h.Token, 3))
			}
		}
		fmt.Printf("✕ No authentication found for host %s. Run polaris auth login to authenticate.\n", insightsHost)
		return nil
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
