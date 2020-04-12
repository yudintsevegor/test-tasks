package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// https://openweathermap.org/api/one-call-api
const urlSchema = "https://api.openweathermap.org/data/2.5/onecall?lat=%v&lon=%v&appid=%v&units=metric"

func (w weatherInformer) getWeatherByAPI(apiKey string, coordinate coordinate) (*response, error) {
	reqUrl := fmt.Sprintf(urlSchema, coordinate.Latitude, coordinate.Longitude, apiKey)
	if w.isDebug {
		log.Printf("reqURL: %s", reqUrl)
	}

	body, err := doRequest(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if w.isDebug {
		log.Printf("body: %s", body)
	}

	response := new(response)
	if err := json.Unmarshal(body, response); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if w.isDebug {
		log.Printf("response: %+v", response)
	}

	return response, nil
}

func doRequest(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("do http requests: %w", err)
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

const layout = "January 2, 15:04 MST"

const textWeather = `
	Time: %v

	Weather: %v
	Temperature: %v °C,
	FeelsLike: %v °C
`

const textAdditionalWeather = `
	Pressure: %v hPa,
	Humidity: %v percent
	Cloudiness: %v percent,
	Visibility: %v m
	Wind speed: %v meter/sec
`

func createWeatherText(weather weather, timeZone string, isCurrentWeather bool) (string, error) {
	standardTime, err := convertUnixTime(weather.DateTime, timeZone)
	if err != nil {
		return "", fmt.Errorf("convert unix time: %w", err)
	}
	weatherDescription := strings.Join(createWeatherDescription(weather.WeatherInfo), ", ")

	out := fmt.Sprintf(textWeather,
		standardTime.Format(layout),
		weatherDescription,
		weather.Temperature,
		weather.FeelsLike,
	)

	if isCurrentWeather {
		out += fmt.Sprintf(textAdditionalWeather,
			weather.Pressure,
			weather.Humidity,
			weather.Clouds,
			weather.Visibility,
			weather.WindSpeed,
		)
	}

	return out, nil
}

func convertUnixTime(unixTimeStamp int64, timeZone string) (time.Time, error) {
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return time.Time{}, fmt.Errorf("load location for time zone: %v: %w", timeZone, err)
	}

	currentTime := time.Unix(unixTimeStamp, 0)

	return currentTime.In(location), nil
}

func createWeatherDescription(weatherInfo []weatherInfo) []string {
	out := make([]string, 0, len(weatherInfo))
	for _, info := range weatherInfo {
		out = append(out, fmt.Sprintf("%s (%s)", info.Main, info.Description))
	}

	return out
}
