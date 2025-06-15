package config

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Token           string
	ChatID          int64
	CronTime        string
	TranslateURL    string
	WeatherURL      string
	MoscowLatitude  float64
	MoscowLongitude float64
	Daily           string
	Hourly          string
	Timezone        string
	ForecastDays    int64
	WindSpeedUnit   string
}

func LoadConfig(logger *slog.Logger) (*Config, error) {
	token := os.Getenv("TOKEN")
	chatIDConfig := os.Getenv("CHAT_ID")
	cronTime := os.Getenv("CRON_TIME")
	translateURL := os.Getenv("TRANSLATE_URL")
	weatherURL := os.Getenv("WEATHER_URL")
	latitudeConfig := os.Getenv("MOSCOW_LATITUDE")
	longitudeConfig := os.Getenv("MOSCOW_LONGITUDE")
	daily := os.Getenv("DAILY")
	hourly := os.Getenv("HOURLY")
	timezone := os.Getenv("TIMEZONE")
	forecastDaysConfig := os.Getenv("FORECAST_DAYS")
	windSpeedUnit := os.Getenv("WIND_SPEED_UNIT")

	if token == "" || chatIDConfig == "" {
		logger.Error("Необходимые переменные окружения отсутствуют", "TOKEN", token, "CHAT_ID", chatIDConfig)
		return nil, errors.New("необходимые переменные окружения отсутствуют")
	}
	if cronTime == "" {
		cronTime = "* 8 * * *"
	}
	chatID, err := strconv.ParseInt(chatIDConfig, 10, 64)
	if err != nil {
		logger.Error("Ошибка преобразования CHAT_ID в int64", "error", err)
		return nil, err
	}
	latitude, err := strconv.ParseFloat(latitudeConfig, 64)
	if err != nil {
		logger.Error("Ошибка преобразования MOSCOW_LATITUDE в float64", "error", err, "MOSCOW_LATITUDE", latitudeConfig)
		return nil, err
	}
	longitude, err := strconv.ParseFloat(longitudeConfig, 64)
	if err != nil {
		logger.Error("Ошибка преобразования MOSCOW_LONGITUDE в float64", "error", err, "MOSCOW_LONGITUDE", longitudeConfig)
		return nil, err
	}
	forecastDays, err := strconv.ParseInt(forecastDaysConfig, 10, 64)
	if err != nil {
		logger.Error("Ошибка преобразования FORECAST_DAYS в int64", "error", err)
		return nil, err
	}

	return &Config{
		Token:           token,
		ChatID:          chatID,
		CronTime:        cronTime,
		TranslateURL:    translateURL,
		WeatherURL:      weatherURL,
		MoscowLatitude:  latitude,
		MoscowLongitude: longitude,
		Daily:           daily,
		Hourly:          hourly,
		Timezone:        timezone,
		ForecastDays:    forecastDays,
		WindSpeedUnit:   windSpeedUnit}, nil

}
