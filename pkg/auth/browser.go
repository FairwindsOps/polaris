package auth

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// openBrowser opens up the provided URL in a browser
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "openbsd":
		fallthrough
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		r := strings.NewReplacer("&", "^&")
		cmd = exec.Command("cmd", "/c", "start", r.Replace(url))
	}
	if cmd != nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			logrus.Printf("Failed to open browser due to error %v", err)
			return fmt.Errorf("failed to open browser: %v", err)
		}
		err = cmd.Wait()
		if err != nil {
			logrus.Printf("Failed to wait for open browser command to finish due to error %v", err)
			return fmt.Errorf("failed to wait for open browser command to finish: %v", err.Error())
		}
		return nil
	} else {
		return errors.New("unsupported platform")
	}
}
