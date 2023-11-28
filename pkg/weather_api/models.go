package weatherapi

const messageModel = "*Forecast for %s, %s*\n_Date:_ %s\n*Weather*\n_Name:_ %s\n_Description:_ %s\n*Atmosphere*\n_Temp:_ %.2f\n_FeelsLike:_ %.2f\n_Pressure:_ %d\n_Humidity:_ %d\n_Wind Speed:_ %.2f\n*Sunrise at* %v\n*Sunset at* %v"

type Forecast struct {
	City    string `json:"name"`
	Date    int    `json:"dt"`
	Weather []struct {
		Name        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Parameters struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Sys struct {
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
}
