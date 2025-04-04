package main

import (
	"log"
	"net/http"
	"os"
	"telegramassist/internal/api"
	"telegramassist/internal/application"
	"telegramassist/internal/infrastructure/mysql"
	"time"

	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

type BotHandler struct {
	esp32Service *application.ESP32Service
	userStates   map[int64]string
	tempData     map[int64]string
}

func NewBotHandler(esp32Service *application.ESP32Service) *BotHandler {
	return &BotHandler{
		esp32Service: esp32Service,
		userStates:   make(map[int64]string),
		tempData:     make(map[int64]string),
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Inicializar repositorio MySQL
	repo, err := mysql.NewMySQLRepository()
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Inicializar servicio
	esp32Service := application.NewESP32Service(repo)

	// Configuración del bot
	pref := tele.Settings{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	handler := NewBotHandler(esp32Service)

	// Comando /start
	b.Handle("/start", func(c tele.Context) error {
		handler.userStates[c.Chat().ID] = ""
		return c.Send("¡Bienvenido! Para registrar tu ESP32, usa uno de los siguientes comandos:\n\n" +
			"/registrar - Registrar un nuevo producto ESP32\n" +
			"/ultimaalerta - Ver la última alerta de tu sensor")
	})

	// Comando /registrar
	b.Handle("/registrar", func(c tele.Context) error {
		handler.userStates[c.Chat().ID] = "waiting_serial"
		return c.Send("Por favor, ingresa el número de serial de tu ESP32:")
	})

	// Comando /ultimaalerta
	b.Handle("/ultimaalerta", func(c tele.Context) error {
		chatID := c.Chat().ID
		reading, err := handler.esp32Service.GetLastKY026ReadingByChat(chatID)
		if err != nil {
			if err.Error() == "No hay un ESP32 registrado para este chat" {
				return c.Send("No tienes ningún ESP32 registrado. Por favor, usa /registrar primero para vincular tu dispositivo.")
			}
			return c.Send("Error al obtener la última lectura: " + err.Error())
		}
		if reading == nil {
			return c.Send("No se encontraron lecturas para tu ESP32.")
		}
		return c.Send("Última lectura del sensor:\n" +
			"Fecha: " + reading.FechaActivacion + "\n" +
			"Estado: " + reading.Estado)
	})

	// Manejador de mensajes de texto
	b.Handle(tele.OnText, func(c tele.Context) error {
		chatID := c.Chat().ID
		state := handler.userStates[chatID]
		text := c.Text()

		switch state {
		case "waiting_serial":
			valid, err := handler.esp32Service.ValidateAndLinkESP32(chatID, text)
			if err != nil {
				return c.Send("Error: " + err.Error())
			}
			if !valid {
				return c.Send("El número de serial no es válido. Por favor, verifica e intenta nuevamente.")
			}
			handler.userStates[chatID] = ""
			return c.Send("¡ESP32 registrado exitosamente! Recibirás alertas cuando se detecte humo o fuego.")

		default:
			return c.Send("Por favor, usa /start para ver los comandos disponibles.")
		}
	})

	log.Println("Bot iniciado...")

	// Inicializar el manejador de alertas
	alertHandler := api.NewAlertHandler(esp32Service, b)

	// Configurar el servidor HTTP
	http.HandleFunc("/api/alerts", alertHandler.HandleAlert)

	// Iniciar el servidor HTTP en una goroutine
	go func() {
		log.Println("Iniciando servidor HTTP en :8080...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
		}
	}()

	// Iniciar el bot
	b.Start()
}
