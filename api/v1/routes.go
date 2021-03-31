package v1

import (
	"github.com/gofiber/fiber/v2"
)

func Routes(route fiber.Router) {
	route.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from JudgeX.")
	})

	route.Get("/languages", GetLanguages)
	route.Post("/judge/:language/:version?", judgeLanguageByVersion)
}
