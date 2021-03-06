package about

import "fmt"

type AboutType struct {
	Version, Email, ProductName, Website string
}

var gitVersion = "local-dev-build"

var About = AboutType{
	ProductName: "Sql Schema Explorer",
	Version:     "0.67-" + gitVersion,
	Website:     "http://schemaexplorer.io/",
	Email:       "tim@schemaexplorer.io",
}

func (about AboutType) Summary() string {
	return fmt.Sprintf("%s v%s %s", about.ProductName, about.Version, about.Website)
}
