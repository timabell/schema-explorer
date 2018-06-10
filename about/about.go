package about

type AboutType struct {
	Version, Email, ProductName, Website string
}

var About = AboutType{
	ProductName: "Sql Schema Explorer",
	Version:     "0.26",
	Website:     "http://schemaexplorer.io/",
	Email:       "sse@timwise.co.uk",
}
