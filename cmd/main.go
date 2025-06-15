package main

import (
	// "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"telegram_bot/internal/config"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func setupLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	return logger
}

type Response struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
}

type ResponseWeather struct {
	Hourly struct {
		Time []string  `json:"time"`
		Temp []float64 `json:"temperature_2m"`
		Rain []float64 `json:"rain"`
		Wind []float64 `json:"wind_speed_10m"`
	} `json:"hourly"`
	Daily struct {
		TempMax []float64 `json:"temperature_2m_max"`
		TempMin []float64 `json:"temperature_2m_min"`
		WindMax []float64 `json:"wind_speed_10m_max"`
		Sunrise []string  `json:"sunrise"`
		Sunset  []string  `json:"sunset"`
	} `json:"daily"`
}

func handleWeather(weather *ResponseWeather) string {
	weatherMap := map[string][]string{}
	for i, value := range weather.Hourly.Time {
		weatherMap[value] = append(weatherMap[value], fmt.Sprintf("%0.1f", weather.Hourly.Temp[i]), fmt.Sprintf("%0.1f", weather.Hourly.Rain[i]), fmt.Sprintf("%0.1f", weather.Hourly.Wind[i]))
	}
	sb := strings.Builder{}

	for _, key := range weather.Hourly.Time {
		time := strings.Split(key, "T")[1]
		if slices.Contains([]string{"09:00", "16:00", "19:00"}, time) {
			time = "*" + strings.Split(key, "T")[1] + "*"
			title := []rune(fmt.Sprintf("*%s*\t\t\t\t🌡️ *+%s*\t\t\t\t\t🌧️ *%s*\t\t\t\t🌬️ *%s*\n", time, weatherMap[key][0], weatherMap[key][1], weatherMap[key][2]))
			sb.WriteString(string(title))
		}
	}
	sunrise := strings.Split(weather.Daily.Sunrise[0], "T")[1]
	sunset := strings.Split(weather.Daily.Sunset[0], "T")[1]
	result := string([]rune("Погода сегодня: \n")) + sb.String()
	result = result + "\n" + string([]rune(fmt.Sprintf("Макс. 🌡️ сегодня: *+%0.1f °C*\nМин. 🌡️ сегодня: *+%0.1f °C*\nМакс. 🌬️ сегодня: *%0.1f м/с*\n\nВосход 🌅: *%s*\nЗакат 🌇: *%s*",
		weather.Daily.TempMax[0], weather.Daily.TempMin[0], weather.Daily.WindMax[0], sunrise, sunset)))
	return result
}

func getWeatherToday(cfg *config.Config, logger *slog.Logger) (string, error) {
	weatherParams := url.Values{}
	weatherParams.Set("latitude", fmt.Sprintf("%f", cfg.MoscowLatitude))
	weatherParams.Set("longitude", fmt.Sprintf("%f", cfg.MoscowLongitude))
	weatherParams.Set("daily", cfg.Daily)
	weatherParams.Set("hourly", cfg.Hourly)
	weatherParams.Set("timezone", cfg.Timezone)
	weatherParams.Set("forecast_days", fmt.Sprintf("%d", cfg.ForecastDays))
	weatherParams.Set("wind_speed_unit", cfg.WindSpeedUnit)

	logger.Info(cfg.WeatherURL + "?" + weatherParams.Encode())

	weatherURL := cfg.WeatherURL + "?" + weatherParams.Encode()

	resp, err := http.Get(weatherURL)
	if err != nil {
		return "", errors.New("ошибка запроса к API при получении погоды")
	}

	defer resp.Body.Close()

	response := ResponseWeather{}
	body, _ := io.ReadAll(resp.Body)
	log.Println(string(body))
	if err := json.Unmarshal(body, &response); err != nil {
		return "", errors.New("ошибка декодирования JSON при переводе текста")
	}

	log.Println(response.Hourly)
	if len(response.Hourly.Rain) == 0 {
		return "", errors.New("пустой ответ от MyMemory при получении погоды")
	}
	if len(response.Hourly.Temp) == 0 {
		return "", errors.New("пустой ответ от MyMemory при получении погоды")
	}
	if len(response.Hourly.Time) == 0 {
		return "", errors.New("пустой ответ от MyMemory при получении погоды")
	}
	if len(response.Hourly.Wind) == 0 {
		return "", errors.New("пустой ответ от MyMemory при получении погоды")
	}

	result := handleWeather(&response)

	return result, nil
}

func translateWord(word string, targetLang string, cfg *config.Config) (string, error) {

	params := url.Values{}
	params.Set("q", word)
	params.Set("langpair", "en|"+targetLang)

	url := cfg.TranslateURL + "?" + params.Encode()

	resp, err := http.Get(url)
	if err != nil {
		return "", errors.New("ошибка запроса к API при переводе текста")
	}
	defer resp.Body.Close()
	response := Response{}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", errors.New("ошибка декодирования JSON при переводе текста")
	}
	log.Println(response.ResponseData.TranslatedText)
	if response.ResponseData.TranslatedText == "" {
		return "", errors.New("пустой ответ от MyMemory при переводе текста")
	}

	translatedText := response.ResponseData.TranslatedText

	return translatedText, nil

}

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Файл .env не найден или не загружен")
	}

	logger := setupLogger()

	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		logger.Error("Не получается создать бота")
	}

	bot.Debug = true

	// u := tgbotapi.NewUpdate(0)
	// u.Timeout = 60

	// updates := bot.GetUpdatesChan(u)

	cn := cron.New()
	defer cn.Stop()

	_, err = cn.AddFunc(cfg.CronTime, func() {

		// for update := range updates {
		// 	if update.Message != nil {
		// 		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		// 		// translatedWord, err := translateWord(update.Message.Text, "ru")
		// 		// if err != nil {
		// 		// 	log.Fatal(err)
		// 		// }
		weather, err := getWeatherToday(cfg, logger)
		if err != nil {
			log.Fatal(err)
		}

		msg := tgbotapi.NewMessage(cfg.ChatID, weather)
		msg.ParseMode = "markdown"
		_, err = bot.Send(msg)
		if err != nil {
			log.Fatal(err)
		}
		// }
		// }
	})
	if err != nil {
		logger.Error("Не удалось добавить cron-задачу")
	}
	cn.Start()
	logger.Info("Планироващик запущен")

	select {}

}
