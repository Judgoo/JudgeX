//go:generate go run ./generate

package languages

import (
	_ "embed"
	"log"

	"gopkg.in/yaml.v2"
)

type LanguageType int

type LanguageRecipe struct {
	Build []string
	Run   []string
}

type LanguageProfile struct {
	Filestem string
	Ext      string
	Filename string
	Output   string
	Build    []string
	Run      string
	Mco      bool
}

type LanguageProfileMap map[string]*LanguageProfile

type LanguageInfo struct {
	Language    *LanguageType
	VersionName string
	Version     *VersionInfo
}

//go:embed languages_impl.yml
var LanguageData []byte

var ProfileMap = new(LanguageProfileMap)

func init() {
	var err = yaml.Unmarshal(LanguageData, ProfileMap)
	if err != nil {
		log.Fatalf("err when load languages: %v", err)
	}
}
