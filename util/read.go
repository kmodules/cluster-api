package util

import (
	"os"
	"io/ioutil"
)

func ReadFile(file string) (string, error)  {
	var err error
	var data []byte
	_, err = os.Stat(file)
	if err == nil {
		data, err = ioutil.ReadFile(file)
		if err == nil {
			return string(data), nil
		}
	}
	return "", err
}
