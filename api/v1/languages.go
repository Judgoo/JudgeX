package v1

import (
	"JudgeX/languages"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetLanguages(c *fiber.Ctx) error {
	return c.SendString(strings.Join(languages.LanguageTypeNames(), ","))
}
