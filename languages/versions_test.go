package languages_test

import (
	"testing"

	"github.com/Judgoo/JudgeX/languages"
	"github.com/stretchr/testify/assert"
)

func TestVersionsExists(t *testing.T) {
	for lang, vs := range languages.VersionNameMap {
		for _, versionName := range vs {
			t.Log("lang", lang)
			t.Log("versionName", versionName)
			versionInfo := languages.VersionInfos[versionName]
			t.Logf("versionInfo: %#v", versionInfo)
			assert.NotNil(t, versionInfo)
		}
	}
}
