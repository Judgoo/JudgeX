package routes

import (
	"fmt"
	"strings"

	"github.com/Judgoo/JudgeX/pkg/api"
	"github.com/Judgoo/JudgeX/pkg/constants"
	"github.com/Judgoo/JudgeX/pkg/entities"
	"github.com/Judgoo/JudgeX/pkg/judge"
	"github.com/Judgoo/languages"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	xUtils "github.com/Judgoo/JudgeX/utils"
)

func JudgeRoutes(route fiber.Router, service judge.Service) {
	route.Post("/judge/:language/:version?", judgeLanguageByVersion(service))
}

func judgeLanguageByVersion(service judge.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		languageString := utils.CopyString(c.Params("language"))
		lt, err := languages.ParseLanguageType(languageString)
		if err != nil {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.LANGUAGE_NOT_FOUND_ERROR, err.Error())
		}

		var requestBody entities.JudgePostData
		err = xUtils.ParseJSONBody(c, &requestBody)
		if err != nil {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.PARSE_BODY_ERROR, err.Error())
		}
		validationErrors := entities.Validate(requestBody)
		if validationErrors != nil {
			return api.ApiAbort(c, fiber.StatusUnprocessableEntity, constants.VALIDATE_ERROR, validationErrors)
		}

		_version := c.Params("version", "")
		versionName, versionInfo, exists := lt.GetVersionInfo(_version)
		if !exists {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.VERSION_NOT_FOUND_ERROR, fmt.Sprintf("version %s not found in language %s", _version, languageString))
		}
		requestid := c.Locals("requestid").(string)
		inputs := requestBody.Inputs
		outputs := requestBody.Outputs
		if len(inputs) == 0 {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.NO_TESTDATA_ERROR, "no testdata found")
		}
		if len(inputs) != len(outputs) {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.TESTDATA_LENGTH_ERROR, "length of inputs and outputs are not equal")
		}
		if strings.TrimSpace(requestBody.Code) == "" {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.CODE_EMPTY_ERROR, "please input code")
		}
		judgeInfo := judge.JudgeInfo{Language: &lt, Version: versionInfo, VersionName: versionName}

		resp, judgeErr := service.Judge(requestid, &requestBody, &judgeInfo)
		if judgeErr != nil {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.SYSTEM_ERROR, judgeErr.Error())
		}
		return api.NormalSuccess(c, resp)
	}
}
