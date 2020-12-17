package main

import (
	"github.com/JustAddRobots/xhplconsole-api/app"
	"github.com/JustAddRobots/xhplconsole-api/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	database, err := db.Conn()
	if err != nil {
		log.Fatalf("Database connection failed: %s", err.Error())
	}

	app := &app.App{
		Router:   mux.NewRouter().StrictSlash(true).PathPrefix("/v1").Subrouter(),
		Database: database,
	}

	app.SetupRouter()
	log.Fatal(http.ListenAndServe(":3456", app.Router))
}
