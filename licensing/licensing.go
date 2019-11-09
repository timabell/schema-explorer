package licensing

import (
	"fmt"
)

var copyrightYear = 2019

func LicenseText() string {
	return fmt.Sprintf("Licensed under A-GPLv3")
}

func CopyrightText() string {
	return fmt.Sprintf("Copyright 2015-%d Tim Abell", copyrightYear)
}
