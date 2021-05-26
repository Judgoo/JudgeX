package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func ReadDataFromFile(file string) ([]byte, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("File reading error:", err)
		return []byte{}, err
	}
	return data, nil
}

type File struct {
	Path    string
	Name    string
	Content []byte
}

func WriteFile(file *File) error {
	err := os.MkdirAll(filepath.Dir(file.Path), os.ModeDir|(OS_USER_RWX|OS_ALL_R))
	if err != nil {
		return errors.Wrapf(err, "create directory %s fail", file.Path)
	}
	return os.WriteFile(file.Path, file.Content, (OS_USER_RW | OS_ALL_R))
}
