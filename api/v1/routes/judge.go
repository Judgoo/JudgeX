package routes

import (
	"fmt"

	"github.com/Judgoo/JudgeX/pkg/api"
	"github.com/Judgoo/JudgeX/pkg/entities"
	"github.com/Judgoo/JudgeX/pkg/languages"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	xUtils "github.com/Judgoo/JudgeX/utils"
)

func JudgeRoutes(route fiber.Router, service languages.Service) {
	route.Post("/judge/:language/:version?", judgeLanguageByVersion(service))
}

func judgeLanguageByVersion(service languages.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		languageString := utils.CopyString(c.Params("language"))
		language, err := languages.ParseLanguageType(languageString)
		if err != nil {
			return api.ApiAbortWithoutData(c, fiber.StatusBadRequest, err.Error())
		}

		var requestBody entities.JudgePostData
		err = xUtils.ParseJSONBody(c, &requestBody)
		if err != nil {
			return api.ApiAbort(c, fiber.StatusBadRequest, "Parse JSON Body Error", err.Error())
		}
		validationErrors := entities.Validate(requestBody)
		if validationErrors != nil {
			return api.ApiAbort(c, fiber.StatusUnprocessableEntity, "Validation Error", validationErrors)
		}
		version := c.Params("version", "")
		id := c.Locals("requestid").(string)

		resp, judgeErr := service.Judge(c, id, &requestBody, &language, version)
		if judgeErr != nil {
			if judgeErr == languages.ErrorLanguageVersionNotFound {
				return api.ApiAbort(c, fiber.StatusBadRequest, judgeErr.Error(), fmt.Sprintf("version %s not found in language %s", version, languageString))
			}
			if judgeErr == languages.ErrorEmptyCode {
				return api.ApiAbort(c, fiber.StatusBadRequest, judgeErr.Error(), "please input code")
			}
			// 这里都看做返回的是 api.ApiAbort
			return judgeErr
		}
		return api.NormalSuccess(c, resp)
	}
}
