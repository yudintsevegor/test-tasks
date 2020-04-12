package main

type response struct {
	Timezone string    `json:"timezone"`
	Current  weather   `json:"current"`
	Hourly   []weather `json:"hourly"`
}

type weather struct {
	DateTime int64 `json:"dt"` // time, unix, UTC

	Temperature float64       `json:"temp"`       // Celsius
	FeelsLike   float64       `json:"feels_like"` // Celsius
	Pressure    float64       `json:"pressure"`   // Atmospheric pressure on the sea level, hPa
	Humidity    float64       `json:"humidity"`   // Humidity, %
	Clouds      float64       `json:"clouds"`     // Cloudiness, %
	Visibility  float64       `json:"visibility"` // Average visibility, meters
	WindSpeed   float64       `json:"wind_speed"` // Wind speed. Unit Default: meter/sec
	WeatherInfo []weatherInfo `json:"weather"`
}

type weatherInfo struct {
	Main        WeatherType `json:"main"`
	Description string      `json:"description"`
}

type WeatherType string

// https://openweathermap.org/weather-conditions#Weather-Condition-Codes-2
const (
	Thunderstorm WeatherType = "Thunderstorm"
	Drizzle      WeatherType = "Drizzle"
	Rain         WeatherType = "Rain"
	Snow         WeatherType = "Snow"
	Clouds       WeatherType = "Clouds"
	Clear        WeatherType = "Clear"
)
