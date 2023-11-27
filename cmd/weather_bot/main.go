package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"git.foxminded.ua/foxstudent106264/task-2.5/cmd/weather_bot/internal/database"
	"git.foxminded.ua/foxstudent106264/task-2.5/pkg/config"
	weatherapi "git.foxminded.ua/foxstudent106264/task-2.5/pkg/weather_api"
	api "git.foxminded.ua/foxstudent106264/tgapi"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type App struct {
	config   config.Config
	tgapi    api.Api
	weather  weatherapi.WeatherAPI
	database database.Database
}

func init() {
	cfg := config.GetConfig()

	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		fmt.Printf("Error parsing log level: %v, setting log level to info\n", err)
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
		fmt.Printf("log level was set to %s\n", cfg.LogLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	fmt.Printf("config initialized\n")
}

func main() {
	log.Info("main app started")
	cfg := config.GetConfig()
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoAddr))
	if err != nil {
		log.Panic(err)
	}

	app := App{
		config: *cfg,
		tgapi: api.Api{
			SendMessageURL: "https://api.telegram.org/bot" + cfg.TelegramToken + "/sendMessage",
			GetUpdatesURL:  "https://api.telegram.org/bot" + cfg.TelegramToken + "/getUpdates",
			HTTPClient:     &http.Client{},
		},
		weather:  *weatherapi.GetWeatherApi(),
		database: database.Database{Db: mongoClient, DbName: "testdb", Collection: "subscriptions"},
	}
	app.tgapi.UserInput = &app.database

	app.tgapi.AddCallback(app.database.HandleStartCommand, "/start")
	app.tgapi.AddCallback(app.database.HandleSubscribeCommand, "/subscribe")
	app.tgapi.AddCallback(app.database.HandleSetTimeCommand, "/settime")
	app.tgapi.AddCallback(app.database.HandleSetLocationCommand, "/setlocation")
	app.tgapi.AddCallback(app.database.HandleUnsubscribeCommand, "/unsubscribe")

	server := &http.Server{
		Addr:              cfg.Port,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           http.HandlerFunc(app.tgapi.HandleTelegramWebHook),
	}

	go app.database.NotifyUser(&app.tgapi, app.weather)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
