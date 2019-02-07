package about

import "fmt"

type AboutType struct {
	Version, Email, ProductName, Website string
}

var About = AboutType{
	ProductName: "Sql Schema Explorer",
	Version:     "0.50",
	Website:     "http://schemaexplorer.io/",
	Email:       "sse@timwise.co.uk",
}

func (about AboutType) Summary() string {
	return fmt.Sprintf("%s v%s %s", about.ProductName, about.Version, about.Website)
}
