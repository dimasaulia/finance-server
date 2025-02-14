package main

import (
	"finance/provider/configuration"
	db "finance/provider/database"
	"finance/provider/http"
	"finance/route"
	"flag"
	"regexp"

	"github.com/go-playground/validator/v10"
)

func main() {
	// Get ENV File
	env := configuration.NewConfiguration(".env").LoadEnv()
	port := env.GetString("SERVER_PORT")
	fork := env.GetBool("SERVER_FORK")

	isManageState := flag.Bool("manage", false, "Operate in management mode. When set to true, the system will perform management related tasks. The default value is false.")
	isMigrate := flag.Bool("migrate", false, "System will perform auto migration")
	flag.Parse()

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
	db := psql.StartPSQLConnection()

	if *isManageState && *isMigrate {
		psql.StartMigration(db)
		return
	}

	// Create Server Instance
	cookieSecret := env.GetString("COOKIE_SECRET")
	server := http.NewHttpServer(port, fork, cookieSecret)
	app := server.Setup()

	// Create Validator Instance

	validation := validator.New()
	validation.RegisterValidation("alphaspace", ValidateAlphaNumSpace)

	// Setup Router
	route.NewRoute(app, db, validation, env).SetupMainRouter()

	server.Start(app)

}

func ValidateAlphaNumSpace(fl validator.FieldLevel) bool {
	var regexAlphaNumSpace = regexp.MustCompile("^[ \\p{L}\\p{N}]+$")
	return regexAlphaNumSpace.MatchString(fl.Field().String())
}
