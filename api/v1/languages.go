package v1

import (
	"JudgeX/languages"

	"github.com/gofiber/fiber/v2"
)

type resultItem struct {
	VersionName string
	Description string
	ExampleCode string
}

func GetLanguages(c *fiber.Ctx) error {
	var result = map[string][]resultItem{}
	for k, v := range languages.VersionMap {
		result[k.String()] = []resultItem{}
		for versionName, _v := range v {
			result[k.String()] = append(result[k.String()], resultItem{
				versionName,
				_v.Description,
				_v.ExampleCode,
			})
		}
	}
	return c.JSON(result)
}
