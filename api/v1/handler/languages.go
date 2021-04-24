package handler

import (
	"fmt"

	"github.com/Judgoo/JudgeX/languages"

	"github.com/gofiber/fiber/v2"
)

type resultItem struct {
	VersionName string `json:"version"`
	DisplayName string `json:"name"`
	Description string `json:"description"`
}

func GetLanguages(c *fiber.Ctx) error {
	var result = map[string][]resultItem{}
	for lang, vs := range languages.VersionNameMap {
		result[lang.String()] = []resultItem{}
		for _, versionName := range vs {
			versionInfo := languages.VersionInfos[versionName]
			result[lang.String()] = append(result[lang.String()], resultItem{
				versionName,
				fmt.Sprintf("%s(%s)", lang.String(), versionInfo.DisplayName),
				versionInfo.Description,
			})
		}
	}
	return c.JSON(result)
}
