package routes

import (
	"github.com/Judgoo/JudgeX/pkg/api"
	"github.com/Judgoo/JudgeX/pkg/constants"
	"github.com/Judgoo/JudgeX/pkg/judge"
	"github.com/Judgoo/languages"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

func LanguageRoutes(route fiber.Router, service judge.Service) {
	route.Get("/languages", getLanguages(service))
	route.Get("/languages/:language", getVersionsByLang(service))
}
func getLanguages(service judge.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		result := service.GetLanguages()
		return api.NormalSuccess(c, result)
	}
}

func getVersionsByLang(service judge.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		languageString := utils.CopyString(c.Params("language"))
		lt, err := languages.ParseLanguageType(languageString)
		if err != nil {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.LANGUAGE_NOT_FOUND_ERROR, err.Error())
		}

		result := service.GetLanguages()
		return api.NormalSuccess(c, result[lt.String()])
	}
}
