package sdv

type aboutType struct {
	Version, Email, ProductName, Website string
}

var About = aboutType{
	ProductName: "Sql Schema Explorer",
	Version:     "0.16",
	Website:     "http://schemaexplorer.io/",
	Email:       "sse@timwise.co.uk",
}
