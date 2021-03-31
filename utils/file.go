package utils

import (
	"fmt"
	"os"
)

func ReadDataFromFile(file string) ([]byte, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("File reading error:", err)
		return []byte{}, err
	}
	return data, nil
}
