package about

import "fmt"

type AboutType struct {
	Version, ProductName, Website string
}

var gitVersion = "local-dev-build"

var About = AboutType{
	ProductName: "Sql Schema Explorer",
	Version:     "0.70-" + gitVersion,
	Website:     "https://github.com/timabell/schema-explorer",
}

func (about AboutType) Summary() string {
	return fmt.Sprintf("%s v%s %s", about.ProductName, about.Version, about.Website)
}
