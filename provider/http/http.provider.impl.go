package http

import "github.com/gofiber/fiber/v2"

type IHttpServer interface {
	Setup() *fiber.App
	Start(*fiber.App)
}
