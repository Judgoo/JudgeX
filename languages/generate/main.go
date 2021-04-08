package main

import (
	"JudgeX/languages"
	"JudgeX/languages/generate/generator"
	"JudgeX/utils"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

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

	profileMap := make(languages.LanguageProfileMap)

	err = yaml.Unmarshal(data, &profileMap)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	keys := make([]string, 0, len(profileMap))
	for k := range profileMap {
		keys = append(keys, k)
	}
	// Make the output more consistent by iterating over sorted keys of map
	sort.Strings(keys)
	for _, k := range keys {
		(*profileMap[k]).Filename = (*profileMap[k]).Filestem + "." + (*profileMap[k]).Ext

		for bIdx, buildString := range (*profileMap[k]).Build {
			tpl := template.New((*profileMap[k]).Filename)
			buf := new(bytes.Buffer)
			tmpl, _ := tpl.Parse(buildString)
			_ = tmpl.Execute(buf, (*profileMap[k]))
			(*profileMap[k]).Build[bIdx] = buf.String()
		}
		tpl := template.New((*profileMap[k]).Filename)
		buf := new(bytes.Buffer)
		tmpl, _ := tpl.Parse((*profileMap[k]).Run)
		_ = tmpl.Execute(buf, (*profileMap[k]))
		(*profileMap[k]).Run = buf.String()
	}

	marshaledBytes, err := yaml.Marshal(profileMap)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	outYmlPath := "languages_impl.yml"
	mode := int(0644)
	err = ioutil.WriteFile(outYmlPath, marshaledBytes, os.FileMode(mode))
	if err != nil {
		fmt.Printf("failed writing to file %s: %s", color.Cyan(outYmlPath), color.Red(err))
	}

	outFilePath := "languages_impl.go"
	g := generator.NewGenerator()
	// Parse the file given in arguments
	raw, err := g.GenerateFromProfile(fileName, &keys, &profileMap)
	if err != nil {
		fmt.Printf("failed generating enums\nInputFile=%s\nError=%s", color.Cyan(fileName), color.RedBg(err))
	}

	err = ioutil.WriteFile(outFilePath, raw, os.FileMode(mode))
	if err != nil {
		fmt.Printf("failed writing to file %s: %s", color.Cyan(outFilePath), color.Red(err))
	}
}
