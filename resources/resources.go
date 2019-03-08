package resources

import (
	"os"
	"path"
)

var basePath = ""

const relativeTemplateFolder = "templates/"

var TemplateFolder = relativeTemplateFolder

func init() {
	// Try relative first, if that doesn't exist switch to absolute.
	// Need both because local build moves the exe but uses build folder working dir (so need relative),
	// but mac double-click doesn't set working folder so need to figure out absolute path to resources
	// If we can't find the relative folder, switch to absolute paths for everything
	// The templates folder is being used as a suitable canary to find out what is working.
	if _, e := os.Stat(relativeTemplateFolder); e != nil && os.IsNotExist(e) {
		exePath, err := os.Executable()
		if err != nil {
			panic("couldn't get exe path for templates/ path: " + err.Error())
		}
		exeFolder := path.Dir(exePath)
		//log.Printf("Using absolute resource paths. Base folder: %s", exeFolder)
		basePath = exeFolder
	}

	// setup individual base paths once to save re-calculating
	if basePath != "" {
		TemplateFolder = path.Join(basePath, relativeTemplateFolder)
	}
}
