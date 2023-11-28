package weatherapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"git.foxminded.ua/foxstudent106264/task-2.5/pkg/config"
	api "git.foxminded.ua/foxstudent106264/tgapi"
	log "github.com/sirupsen/logrus"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

type WeatherAPI struct {
	RequestURL string
	HTTPClient
}

var once sync.Once

var weatherApiInstance *WeatherAPI

func GetWeatherApi() *WeatherAPI {
	if weatherApiInstance == nil {
		once.Do(func() {
			fmt.Println("Creating WeatherAPI instance now.")

			cfg := config.GetConfig()

			weatherApiInstance = &WeatherAPI{"https://api.openweathermap.org/data/2.5/weather?APPID=" + cfg.WeatherToken + "&lat=%.2f&lon=%.2f", &http.Client{}}
		})

	}

	return weatherApiInstance
}

func (w *WeatherAPI) GetForecast(latitude, longitude float64) (string, error) {

	resp, getErr := w.Get(fmt.Sprintf(w.RequestURL, latitude, longitude))

	if getErr != nil || (resp.StatusCode != 200) {
		return "", fmt.Errorf("error while making request to weather api: %w", getErr)
	}
	defer resp.Body.Close()

	var forecast Forecast

	if err := json.NewDecoder(resp.Body).Decode(&forecast); err != nil {
		return "", fmt.Errorf("could not decode response from weather api: %w", err)
	}

	date := time.Unix(int64(forecast.Date), 0).UTC().Format(time.DateOnly)
	sunrise := time.Unix(int64(forecast.Sys.Sunrise), 0).UTC().Format(time.TimeOnly)
	sunset := time.Unix(int64(forecast.Sys.Sunset), 0).UTC().Format(time.TimeOnly)

	message := fmt.Sprintf(messageModel,
		forecast.City, forecast.Sys.Country, date, forecast.Weather[0].Name,
		forecast.Weather[0].Description, forecast.Parameters.Temp, forecast.Parameters.FeelsLike, forecast.Parameters.Pressure,
		forecast.Parameters.Humidity, forecast.Wind.Speed, string(sunrise), string(sunset))
	return message, nil
}

func (w *WeatherAPI) LocationWebhook(a *api.Api, update api.Update) {
	weather, err := w.GetForecast(update.Message.Location.Latitude, update.Message.Location.Longitude)
	if err != nil {
		log.Warnf("%s", err)
	}

	a.SendMessageWithLog(weather, update.Message.Chat.Id)

}

func (w *WeatherAPI) StartWebhook(a *api.Api, update api.Update) {

	a.SendMessageWithLog("*Please send your location to get the weather forecast*", update.Message.Chat.Id)

}
