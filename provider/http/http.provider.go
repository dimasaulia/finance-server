package http

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type HttpServerImpl struct {
	Port string
	Fork bool
}

func NewHttpServer(port string, fork bool) IHttpServer {
	return &HttpServerImpl{
		Port: port,
		Fork: fork,
	}
}

func (h *HttpServerImpl) Setup() *fiber.App {
	var appConfig fiber.Config = fiber.Config{
		Prefork: h.Fork,
	}

	app := fiber.New(appConfig)

	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		// For more options, see the Config section
		Format:     "${time} ${ip}:${port} ${status} - ${method} ${path}\n",
		TimeZone:   "Asia/Jakarta",
		TimeFormat: "02/01/2006 15:04:05.000000",
		Output:     io.Writer(os.Stdout),
	}))

	app.Use(func(ctx *fiber.Ctx) error {
		// Middleware to handle panic
		defer func() {
			// Auto recovery process
			if r := recover(); r != nil {
				// Check if new error match custom default
				if errDefault, okDefault := r.(error); okDefault {
					var errString string = "Internal Server Error"
					if errDefault.Error() != "" {
						errString = errDefault.Error()
					}

					ctx.Status(fiber.StatusInternalServerError).JSON(
						fiber.Map{
							"message": errString,
							"success": false,
						},
					)
					return
				}

				// Default handler
				ctx.Status(fiber.StatusInternalServerError).JSON(
					fiber.Map{
						"message": r,
						"success": false,
					},
				)
				return
			}
		}()

		// If no error appear continue process
		return ctx.Next()
	})

	return app
}

func (h *HttpServerImpl) Start(app *fiber.App) {
	port := fmt.Sprintf(":%s", h.Port)
	fmt.Printf("Server Runing On Port: %s \n", h.Port)
	err := app.Listen(port)
	if err != nil {
		log.Fatal(err)
	}
}
