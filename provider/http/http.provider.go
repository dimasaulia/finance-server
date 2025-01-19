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
