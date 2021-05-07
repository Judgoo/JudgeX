package api

import (
	"github.com/gofiber/fiber/v2"
)

func Success(c *fiber.Ctx, code int, message string, data interface{}) error {
	return c.Status(200).JSON(Response{Code: code, Message: message, Data: &data})
}

func NormalSuccess(c *fiber.Ctx, data interface{}) error {
	return c.Status(200).JSON(Response{Code: 200, Message: "success", Data: &data})
}
