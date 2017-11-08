package sdv

type aboutType struct {
	Version, Email, ProductName, Website string
}

var About = aboutType{
	ProductName: "Sql Data Viewer",
	Version:     "0.8",
	Website:     "https://sqldataviewer.com/",
	Email:       "sdv@timwise.co.uk",
}
