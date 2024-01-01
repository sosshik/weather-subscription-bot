package weatherapi

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sosshik/weather-subscription-bot/pkg/weather_api/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetForecast(t *testing.T) {
	// Mock HTTP client
	httpClient := mocks.NewHTTPClient(t)
	weatherApi := GetWeatherApi()
	weatherApi.HTTPClient = httpClient

	// Test cases
	testCases := []struct {
		name           string
		latitude       float64
		longitude      float64
		mockResponse   *http.Response
		mockErr        error
		expectedResult string
		expectedError  string
	}{
		{
			name:           "Success",
			latitude:       36.13,
			longitude:      49.88,
			mockResponse:   &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(mockRespStr))},
			mockErr:        nil,
			expectedResult: exp,
			expectedError:  "",
		},
		{
			name:           "HTTP Client Error",
			latitude:       37.7749,
			longitude:      -122.4194,
			mockResponse:   &http.Response{StatusCode: 404, Body: http.NoBody},
			mockErr:        io.EOF,
			expectedResult: "",
			expectedError:  "could not decode response from weather api: EOF",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			httpClient.On("Get", mock.Anything).Return(tc.mockResponse, tc.mockErr)

			result, err := weatherApi.GetForecast(tc.latitude, tc.longitude)

			if tc.expectedError != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}

			httpClient.AssertExpectations(t)
		})
	}
}

var exp = "*Forecast for Vysokyy, UA*\n_Date:_ 2023-11-23\n*Weather*\n_Name:_ Clouds\n_Description:_ overcast clouds\n*Atmosphere*\n_Temp:_ 266.35\n_FeelsLike:_ 259.35\n_Pressure:_ 1010\n_Humidity:_ 84\n_Wind Speed:_ 6.72\n*Sunrise at* 04:59:44\n*Sunset at* 13:44:07"
var mockRespStr = `{"coord":{"lon":36.13,"lat":49.88},"weather":[{"id":804,"main":"Clouds","description":"overcast clouds","icon":"04n"}],"base":"stations","main":{"temp":266.35,"feels_like":259.35,"temp_min":266.35,"temp_max":266.35,"pressure":1010,"humidity":84,"sea_level":1010,"grnd_level":985},"visibility":10000,"wind":{"speed":6.72,"deg":213,"gust":15.03},"clouds":{"all":100},"dt":1700750175,"sys":{"country":"UA","sunrise":1700715584,"sunset":1700747047},"timezone":7200,"id":688696,"name":"Vysokyy","cod":200}`
