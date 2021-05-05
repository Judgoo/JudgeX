package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

func ApiAbort(c *fiber.Ctx, code int, message string, data interface{}) error {
	if message == "" {
		if desp := utils.StatusMessage(code); desp != "" {
			message = desp
		}
	}
	return c.Status(code).JSON(Response{Code: code, Message: message, Data: &data})
}

func ApiAbortWithoutData(c *fiber.Ctx, code int, message string) error {
	return ApiAbort(c, code, message, nil)
}
