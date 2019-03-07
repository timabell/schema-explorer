package browser

import (
	"log"
	"os/exec"
)

func LaunchBrowser(url string) {
	err := exec.Command(browserCommand, url).Run()
	if err != nil {
		log.Printf("Failed to launch browser automatically: %s", err)
	}
}
