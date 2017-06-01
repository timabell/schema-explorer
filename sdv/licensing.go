package sdv

import (
	"time"
	"log"
	"fmt"
)

// roughly 3 months from when this is released into the wild
var Expiry = time.Date(2017, time.July, 1, 0, 0, 0, 0, time.UTC)
var CopyrightYear = 2017

func Licensing() {
	if time.Now().After(Expiry) {
		log.Panic("Expired trial, contact sdv@timwise.co.uk to obtain a license")
	}
}

func CopyrightText() (string) {
	return fmt.Sprintf("Sql Data Viewer v%s; Copyright 2015-%d Tim Abell <sdv@timwise.co.uk>", Version, CopyrightYear)
}
