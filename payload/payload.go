package payload

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhoonb/archivex"
)

//Payload is a struct
type Payload struct {
	Filename            string
	Path                string
	DeleteAfterTransfer bool
}

// Delete is a function delete payload absalute path from a disk
func (p *Payload) Delete() error {
	return os.RemoveAll(p.Path)
}

// FromArgs returns a payload from args
func FromArgs(args []string) (Payload, error) {
	shouldzip := len(args) > 1
	var files []string
	// Check if content exists
	for _, arg := range args {
		file, err := os.Stat(arg)
		if err != nil {
			return Payload{}, err
		}
		// If at least one argument is dir, the content will be zipped
		if file.IsDir() {
			shouldzip = true
		}
		files = append(files, arg)
	}
	// Prepare the content
	// TODO: Research cleaner code
	var content string
	content = args[0]

	//delet file except README.md
	shouldzip = strings.Index(strings.Replace(content, " ", "", -1), "README.md") <= -1
	return Payload{
		Path:                content,
		Filename:            filepath.Base(content),
		DeleteAfterTransfer: shouldzip,
	}, nil
}

// ZipFiles and return the resulting zip's filename
func ZipFiles(files []string) (string, error) {
	zip := new(archivex.ZipFile)
	tmpfile, err := ioutil.TempFile("", "qrcp")
	if err != nil {
		return "", err
	}
	tmpfile.Close()
	if err := os.Rename(tmpfile.Name(), tmpfile.Name()+".zip"); err != nil {
		return "", err
	}
	zip.Create(tmpfile.Name() + ".zip")
	for _, filename := range files {
		fileinfo, err := os.Stat(filename)
		if err != nil {
			return "", err
		}
		if fileinfo.IsDir() {
			zip.AddAll(filename, true)
		} else {
			file, err := os.Open(filename)
			if err != nil {
				return "", err
			}
			defer file.Close()
			if err := zip.Add(filename, file, fileinfo); err != nil {
				return "", err
			}
		}
	}
	if err := zip.Close(); err != nil {
		return "", nil
	}
	return zip.Name, nil
}
