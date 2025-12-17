package utils

import (
	"encoding/json"
	"os"
)

func LoadArrayMsgs(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return  nil, err
	}
	defer file.Close()

	var arrayMsgs []string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&arrayMsgs); err != nil {
		return  nil, err
	}
	return arrayMsgs, nil
}