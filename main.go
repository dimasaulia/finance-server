package main

import (
	"finance/provider/configuration"
	"finance/provider/http"
	"finance/route"
)

func main() {
	// Get ENV File
	env := configuration.NewConfiguration(".env").LoadEnv()
	port := env.GetString("SERVER_PORT")
	fork := env.GetBool("SERVER_FORK")

	// Create Server Instance
	server := http.NewHttpServer(port, fork)
	app := server.Setup()

	// Setup Router
	route.NewRoute(app).SetupMainRouter()

	server.Start(app)

}
