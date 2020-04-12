## NOTE: see this [repository](https://github.com/yudintsevegor/wheather_informer), please. i took a mistake

# [Умный сервис прогноза погоды (Задача со звездочкой)](https://www.notion.so/03f6716315e04acea3023766e5f2cc0e)

## Architecture

### Language and tools
* [Go](https://golang.org/)
* [Wrapper for Tg-Bot-API](https://gopkg.in/telegram-bot-api.v4)
* [heroku](https://dashboard.heroku.com/) for  deployment

### UI
You can use telegram-bot by [url](https://t.me/informer_weather_bot) or find it by `@informer_weather_bot`

### Request to Bot

Request to the bot consist of one string in format:
```
    Country-City or Country-State-City
```

Country format: iso-2-alpha
State format: two letters
City format: full name of the city

Example:
````
    ru-moscow or us-il-chicago
````

NOTE: register of letters doesnt matter

### Response 

Response consist of 3 part:

* current weather in region
* weather for the next 12 hours
* recommendations

current weather in region

````
 Time: April 12, 09:32 CDT

 Weather: Clouds (overcast clouds)
 Temperature: 11.9 °C,
 FeelsLike: 8.06 °C

 Pressure: 1006 hPa,
 Humidity: 66 percent
 Cloudiness: 90 percent,
 Visibility: 16093 m
 Wind speed: 4.1 meter/sec

````

weather for the next 12 hours in format

````
 Time: April 12, 09:00 CDT

 Weather: Clouds (overcast clouds)
 Temperature: 11.9 °C,
 FeelsLike: 8.48 °C
````

recommendations

````
    Recommendations: 
    Take trousers and jacket. Clouds are outside.Take coat and umbrella. It's Rain outside.
````

### Step-By-Step

#### Initial
* Creation of map with key: location in format `Country-State-City` or `Country-City` and value: coordinates for APi
 request to [weather API](https://openweathermap.org/api)
 
* Load configuration from config.env:
```
    BOT_TOKEN=
    WEB_HOOK=
    WEATHER_KEY_API=
    
    DEBUG=false
```

* Bot creation by `BOT_TOKEN` and setting `WEB_HOOK`

* Create structure for comfortable continues work and start bot 

* Setting handler func for [path](https://weatherinformer.herokuapp.com/cities) with supporting locations

#### Start Bot

* Listen For Web Hook

* Read channel with messages from users

* Handle message and send message to user 


#### Handle message

* Handle text message and get `lat, lon` from map [created on the Initial Step]
* Using [weather API](https://openweathermap.org/api) send request with `lon, lat and api key`
* Create text for user
* Create recommendations for user 

### Additional
I use [samples](http://bulk.openweathermap.org/sample/) for getting information about locations

### Usage

* git clone current repository

* Create config/config.env with  your variables

```
    BOT_TOKEN=
    WEB_HOOK=
    WEATHER_KEY_API=
    
    DEBUG=false
```

* in Makefile change `DEFAULT_PORT`, if you want. And do

````
    make all
````

### Example

You can watch a video with usage [video](https://cloud.mail.ru/public/3rVy/2SyNqqFMd)

### TODO

* Add database for dialog with user
* Add opportunity to ask only current weather/hourly weather for period/etc