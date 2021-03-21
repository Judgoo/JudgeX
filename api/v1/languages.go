package v1

import (
	pkg "JudgeX/pkg"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func languages(c *fiber.Ctx) error {
	return c.SendString(strings.Join(pkg.LanguageTypeNames(), ","))
}
