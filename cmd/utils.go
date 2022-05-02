package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func CreateDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	return err
}

func CreateFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

func CheckErr(msg interface{}) {
	if msg != nil {
		log.Fatal(msg)
		os.Exit(1)
	}
}
