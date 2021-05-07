package routes

import (
	"github.com/Judgoo/JudgeX/pkg/api"
	"github.com/Judgoo/JudgeX/pkg/judge"
	"github.com/gofiber/fiber/v2"
)

func LanguageRoutes(route fiber.Router, service judge.Service) {
	route.Get("/languages", getLanguages(service))
}
func getLanguages(service judge.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		result := service.GetLanguages()
		return api.NormalSuccess(c, result)
	}
}
