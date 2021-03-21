package utils

import (
	"github.com/gofiber/fiber/v2"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

var JSON = jsoniter.Config{
	EscapeHTML:              false,
	MarshalFloatWith6Digits: true,
}.Froze()

func ParseJSONBody(c *fiber.Ctx, out interface{}) error {
	extra.RegisterFuzzyDecoders()
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)

	if c.Is("json") {
		return JSON.Unmarshal(c.Request().Body(), out)
	}
	return fiber.ErrUnprocessableEntity
}

func JSONMarshal(v interface{}) ([]byte, error) {
	return JSON.Marshal(v)
}
