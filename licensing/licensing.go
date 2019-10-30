package licensing

import (
	"github.com/timabell/schema-explorer/about"
	"fmt"
	"log"
	"time"
)

// at least 6 months from when this is released into the wild
var expiry = time.Date(2020, time.July, 2, 0, 0, 0, 0, time.UTC)
var copyrightYear = 2019

func EnforceLicensing() {
	if time.Now().After(expiry) {
		log.Panicf("Expired trial, contact %s to obtain a license", about.About.Email)
	}
}

func LicenseText() string {
	return fmt.Sprintf("This pre-release software will expire on: %s, contact %s for a license.", expiry, about.About.Email)
}

func CopyrightText() string {
	return fmt.Sprintf("Copyright 2015-%d Tim Abell", copyrightYear)
}
