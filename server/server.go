package server

import (
	"JudgeX/pkg"
	xUtils "JudgeX/utils"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/helmet/v2"
)

func setupMiddlewares(app *fiber.App) {
	app.Use(helmet.New())
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 1
	}))
	app.Use(etag.New())
	app.Use(logger.New())
	app.Use(requestid.New())
}

func registerBuiltinRoutes(app *fiber.App) {
	app.Get("/stack", func(c *fiber.Ctx) error {
		return c.JSON(c.App().Stack())
	})
}

func Create() *fiber.App {
	app := fiber.New(fiber.Config{
		Prefork:      false,
		ServerHeader: "JudgeX 0.0.1",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*fiber.Error); ok {
				return pkg.ApiAbort(c, e.Code, e.Message, e.Error())
			} else {
				return pkg.ApiAbortWithoutData(c, 500, err.Error())
			}
		},
		JSONEncoder: xUtils.JSONMarshal,
	})

	setupMiddlewares(app)
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
