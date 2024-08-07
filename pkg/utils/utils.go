package utils

import (
	"archive/zip"
	"io"
	"strings"
)

func FindFileInZipUseCallback(filePath string, targetFunc func(io.ReadCloser)) {
	// Open the zip file
	r, err := zip.OpenReader(filePath)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	// Iterate through the files in the archive, find the xml file and open it
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			if err != nil {
				panic(err)
			}
			defer rc.Close()

			// Do something with the xml file
			targetFunc(rc)
		}
	}
}
