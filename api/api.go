package api

import (
	v1 "JudgeX/api/v1"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	v1Route := app.Group("/v1", func(c *fiber.Ctx) error {
		c.Set("X-Judge-Version", "v1")
		return c.Next()
	})
	v1.Routes(v1Route)
}
