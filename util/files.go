package util

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Exists returns whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Lstat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// CreateFolder creates a new directory for the given path
func CreateFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteToFile(data []byte, outFile string) error {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

// writes a given struct to the given outFile
func WriteStructToFile(class interface{}, outFile string) {
	if !Exists(outFile) {
		if err := os.MkdirAll(filepath.Dir(outFile), 0770); err != nil {
			logrus.Println("Could not create path!")
		}
	}

	d, err := yaml.Marshal(&class)
	if err != nil {
		logrus.Fatalf("error: %v", err)
	}

	_, err = os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Println("Could not open file", outFile)
	}
	err = ioutil.WriteFile(outFile, d, 0644)
	if err != nil {
		logrus.Println("Could not write file", outFile)
	}
}
