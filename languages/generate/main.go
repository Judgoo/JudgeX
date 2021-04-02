package main

import (
	"JudgeX/languages"
	"JudgeX/languages/generate/generator"
	"JudgeX/utils"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/labstack/gommon/color"
)

func main() {
	fileName, _ := filepath.Abs("languages.go")
	profilePath, _ := filepath.Abs("languages.yml")
	data, err := utils.ReadDataFromFile(profilePath)
	if err != nil {
		os.Exit(1)
	}

	m := make(languages.LanguageProfileMap)
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// fmt.Printf("%#v", m)
	outFilePath := "languages_impl.go"
	g := generator.NewGenerator()
	// Parse the file given in arguments
	raw, err := g.GenerateFromProfile(fileName, &m)
	if err != nil {
		fmt.Printf("failed generating enums\nInputFile=%s\nError=%s", color.Cyan(fileName), color.RedBg(err))
	}

	mode := int(0644)
	err = ioutil.WriteFile(outFilePath, raw, os.FileMode(mode))
	if err != nil {
		fmt.Printf("failed writing to file %s: %s", color.Cyan(outFilePath), color.Red(err))
	}
}
