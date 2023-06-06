package auth

import "fmt"

func PrintToken() error {
	if content, err := readPolarisHostsFile(); err == nil {
		if len(content) > 0 {
			for k, h := range content {
				if len(content) == 1 {
					fmt.Println(h.Token)
				} else {
					fmt.Printf("%s: %s\n", k, h.Token)
				}
			}
			return nil
		}
	}
	fmt.Println("no oauth token")
	return nil
}
