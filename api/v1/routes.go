package v1

import (
	"JudgeX/api/v1/handler"

	"github.com/gofiber/fiber/v2"
)

func Routes(route fiber.Router) {
	route.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from JudgeX.")
	})

	route.Get("/languages", handler.GetLanguages)
	route.Post("/judge/:language/:version?", handler.JudgeLanguageByVersion)
}
