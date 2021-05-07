package routes

import (
	"github.com/Judgoo/JudgeX/pkg/api"
	"github.com/Judgoo/JudgeX/pkg/languages"
	"github.com/gofiber/fiber/v2"
)

func LanguageRoutes(route fiber.Router, service languages.Service) {
	route.Get("/languages", getLanguages(service))
}
func getLanguages(service languages.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		result := service.GetLanguages()
		return api.NormalSuccess(c, result)
	}
}
