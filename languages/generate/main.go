package main

import (
	"JudgeX/languages/generate/generator"
	"JudgeX/utils"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/labstack/gommon/color"
	"gopkg.in/yaml.v3"
)

// 一些计算规则
// FILE = filestem + '.' + ext
// OUTPUT = output

type LanguageProfile struct {
	Filestem string
	Output   string
	Ext      string
	Build    []string
	Run      []string
	Mco      bool
}
type LanguageProfiles struct {
	Assembly     LanguageProfile
	Bash         LanguageProfile
	C            LanguageProfile
	Clojure      LanguageProfile
	CoffeeScript LanguageProfile
	Cpp          LanguageProfile
	CSharp       LanguageProfile
	D            LanguageProfile
	Elixir       LanguageProfile
	Go           LanguageProfile
	Groovy       LanguageProfile
	Haskell      LanguageProfile
	Java         LanguageProfile
	JavaScript   LanguageProfile
	Julia        LanguageProfile
	Kotlin       LanguageProfile
	Lua          LanguageProfile
	Nim          LanguageProfile
	Ocaml        LanguageProfile
	Perl         LanguageProfile
	Php          LanguageProfile
	Python       LanguageProfile
	Raku         LanguageProfile
	Ruby         LanguageProfile
	Rust         LanguageProfile
	Scala        LanguageProfile
	Swift        LanguageProfile
	TypeScript   LanguageProfile
	Unknown      LanguageProfile
}

func main() {
	fileName, _ := filepath.Abs("languages.go")
	profilePath, _ := filepath.Abs("languages.yml")
	data, err := utils.ReadDataFromFile(profilePath)
	if err != nil {
		os.Exit(1)
	}
	lang := LanguageProfiles{}
	err = yaml.Unmarshal(data, &lang)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", lang)
	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m:\n%v\n\n", m)

	outFilePath := "languages_impl.go"
	g := generator.NewGenerator()
	g.WithMarshal()
	g.WithNames()
	g.WithNoPrefix()
	g.WithCaseInsensitiveParse()
	// Parse the file given in arguments
	raw, err := g.GenerateFromFile(fileName)
	if err != nil {
		fmt.Printf("failed generating enums\nInputFile=%s\nError=%s", color.Cyan(fileName), color.RedBg(err))
	}

	mode := int(0644)
	err = ioutil.WriteFile(outFilePath, raw, os.FileMode(mode))
	if err != nil {
		fmt.Printf("failed writing to file %s: %s", color.Cyan(outFilePath), color.Red(err))
	}
}
