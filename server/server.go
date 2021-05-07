package server

import (
	"fmt"
	"os"

	v1 "github.com/Judgoo/JudgeX/api/v1/routes"
	"github.com/Judgoo/JudgeX/pkg/api"
	"github.com/Judgoo/JudgeX/pkg/flake"
	"github.com/Judgoo/JudgeX/pkg/judge"
	xUtils "github.com/Judgoo/JudgeX/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func setupMiddlewares(app *fiber.App) {
	app.Use(recover.New())
	app.Use(requestid.New(requestid.Config{
		Next:       nil,
		Header:     fiber.HeaderXRequestID,
		Generator:  flake.Digest,
		ContextKey: "requestid",
	}))
}

func registerRoutes(app *fiber.App) {
	v1Route := app.Group("/v1", func(c *fiber.Ctx) error {
		c.Set("X-Judge-Version", "v1")
		return c.Next()
	})
	judgeService := judge.NewService()
	v1.JudgeRoutes(v1Route, judgeService)
	v1.LanguageRoutes(v1Route, judgeService)
}

func registerBuiltinRoutes(app *fiber.App) {
	app.Get("/stack", func(c *fiber.Ctx) error {
		return c.JSON(c.App().Stack())
	})
}

func Create() *fiber.App {
	app := fiber.New(fiber.Config{
		Prefork:      true,
		ServerHeader: "JudgeX 0.0.1",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*fiber.Error); ok {
				return api.ApiAbort(c, e.Code, e.Message, e.Error())
			} else {
				return api.ApiAbortWithoutData(c, 500, err.Error())
			}
		},
		JSONEncoder: xUtils.JSONMarshal,
	})

	setupMiddlewares(app)
	registerRoutes(app)
	registerBuiltinRoutes(app)

	app.Get("/", func(c *fiber.Ctx) error {
		// TODO render server info
		return c.SendString("OK")
	})

	return app
}

func Listen(app *fiber.App) error {
	// add a middleware function at the very bottom of the stack
	// (below all other functions) to handle a 404 response:
	app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404)
	})

	serverHost, ok1 := os.LookupEnv("SERVER_HOST")
	if !ok1 {
		serverHost = "0.0.0.0"
	}
	serverPort, ok2 := os.LookupEnv("SERVER_PORT")
	if !ok2 {
		serverPort = "3000"
	}

	return app.Listen(fmt.Sprintf("%s:%s", serverHost, serverPort))
}

func Release() {
	judge.Release()
}
