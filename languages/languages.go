//go:generate go run ./generate

package languages

import (
	_ "embed"

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

//go:embed languages_impl.yml
var LanguageData []byte
var ProfileMap = make(LanguageProfileMap)
var _ = yaml.Unmarshal(LanguageData, &ProfileMap)

func (lang LanguageType) Profile() *LanguageProfile {
	return ProfileMap[lang.String()]
}

var OnlyCheckMem = []LanguageType{Java, Kotlin, JavaScript, TypeScript}
