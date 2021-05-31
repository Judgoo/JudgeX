package routes

import (
	"fmt"
	"strings"

	"github.com/Judgoo/JudgeX/pkg/api"
	"github.com/Judgoo/JudgeX/pkg/constants"
	"github.com/Judgoo/JudgeX/pkg/entities"
	"github.com/Judgoo/JudgeX/pkg/judge"
	rEntities "github.com/Judgoo/Judger/entities"
	"github.com/Judgoo/languages"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	xUtils "github.com/Judgoo/JudgeX/utils"
)

func JudgeRoutes(route fiber.Router, service judge.Service) {
	route.Post("/judge/:language/:version_id?", judgeLanguageByVersion(service, false))
	route.Post("/judgex/:language/:version_id?", judgeLanguageByVersion(service, true))
}

func judgeLanguageByVersion(service judge.Service, isJudgeX bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		languageString := utils.CopyString(c.Params("language"))
		lt, err := languages.ParseLanguageType(languageString)
		if err != nil {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.LANGUAGE_NOT_FOUND_ERROR, err.Error())
		}

		var requestBody rEntities.JudgePostData
		err = xUtils.ParseJSONBody(c, &requestBody)
		if err != nil {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.PARSE_BODY_ERROR, err.Error())
		}
		validationErrors := entities.Validate(requestBody)
		if validationErrors != nil {
			return api.ApiAbort(c, fiber.StatusUnprocessableEntity, constants.VALIDATE_ERROR, validationErrors)
		}

		versionId := c.Params("version_id", "")
		versionName, versionInfo, exists := lt.GetVersionInfo(versionId)
		if !exists {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.VERSION_NOT_FOUND_ERROR, fmt.Sprintf("version id(%s) not found in language %s", versionId, languageString))
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
		judgeInfo := rEntities.JudgeInfo{Language: &lt, Version: versionInfo, VersionName: versionName}
		var resp *judge.JudgeResponse
		var judgeErr error
		if isJudgeX {
			resp, judgeErr = service.JudgeX(requestid, &requestBody, &judgeInfo)
		} else {
			resp, judgeErr = service.Judge(requestid, &requestBody, &judgeInfo)
		}
		if judgeErr != nil {
			return api.ApiAbort(c, fiber.StatusBadRequest, constants.SYSTEM_ERROR, judgeErr.Error())
		}
		return api.NormalSuccess(c, resp)
	}
}
