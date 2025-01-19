package http

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
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
