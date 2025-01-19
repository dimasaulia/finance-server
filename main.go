package main

import (
	"finance/provider/configuration"
	db "finance/provider/database"
	"finance/provider/http"
	"finance/route"
)

func main() {
	// Get ENV File
	env := configuration.NewConfiguration(".env").LoadEnv()
	port := env.GetString("SERVER_PORT")
	fork := env.GetBool("SERVER_FORK")

	// Start DB Connection
	dbConf := db.PSQLConfiguration{
		Host:     env.GetString("DB_HOST"),
		User:     env.GetString("DB_USER"),
		Password: env.GetString("DB_PASSWORD"),
		Name:     env.GetString("DB_NAME"),
		Port:     env.GetString("DB_PORT"),
		SSL:      env.GetString("DB_SSL"),
	}

	psql := db.NewPSQLConnetion(dbConf)
	psql.StartPSQLConnection()

	// Create Server Instance
	server := http.NewHttpServer(port, fork)
	app := server.Setup()

	// Setup Router
	route.NewRoute(app).SetupMainRouter()

	server.Start(app)

}
