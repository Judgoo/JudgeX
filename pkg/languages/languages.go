//go:generate go run ./generate

package languages

import (
	_ "embed"
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
