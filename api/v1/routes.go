package v1

import (
	pkg "JudgeX/pkg"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

func Routes(route fiber.Router) {
	route.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("JudgeX v1")
	})

	route.Post("/judge/:language/:version", func(c *fiber.Ctx) error {
		language := utils.CopyString(c.Params("language"))
		version := utils.CopyString(c.Params("version"))
		language_enum, err := pkg.ParseLanguageType(language)
		if err != nil {
			return c.SendString(err.Error())
		}

		return c.SendString("Hello, World!" + language_enum.String() + version)
	})

	route.Post("/judge/:languagesId", func(c *fiber.Ctx) error {
		languages := utils.CopyString(c.Params("languages"))
		return c.SendString("Hello, World!" + languages)
	})

	route.Get("/languages", func(c *fiber.Ctx) error {
		return c.SendString("ssss")
	})
}
