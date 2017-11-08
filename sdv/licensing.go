package sdv

import (
	"fmt"
	"log"
	"time"
)

// roughly 3 months from when this is released into the wild
var Expiry = time.Date(2018, time.February, 1, 0, 0, 0, 0, time.UTC)
var CopyrightYear = 2017

func Licensing() {
	if time.Now().After(Expiry) {
		log.Panic("Expired trial, contact sdv@timwise.co.uk to obtain a license")
	}
}

func CopyrightText() string {
	return fmt.Sprintf("Copyright 2015-%d Tim Abell", CopyrightYear)
}
