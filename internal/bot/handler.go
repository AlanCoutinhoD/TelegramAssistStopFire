package bot

import (
    "telegramassist/internal/application"
    tele "gopkg.in/telebot.v3"
    "os"
    "time"
    "log"
	strconv "strconv"
)

type BotHandler struct {
    esp32Service *application.ESP32Service
    ky026Service *application.KY026Service
    Bot          *tele.Bot
    userStates   map[int64]string
    tempData     map[int64]string
}

func NewBotHandler(esp32Service *application.ESP32Service, ky026Service *application.KY026Service) *BotHandler {
    return &BotHandler{
        esp32Service: esp32Service,
        ky026Service: ky026Service,
        userStates:   make(map[int64]string),
        tempData:     make(map[int64]string),
    }
}

func (h *BotHandler) HandleUltimaAlerta(c tele.Context) error {
    chatID := c.Chat().ID
  
    reading, err := h.ky026Service.GetLastReading(strconv.FormatInt(chatID, 10))
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
}

func (h *BotHandler) Start() {
    pref := tele.Settings{
        Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
        Poller: &tele.LongPoller{Timeout: 10 * time.Second},
    }

    bot, err := tele.NewBot(pref)
    if err != nil {
        log.Fatal(err)
        return
    }
    h.Bot = bot

	
    h.setupCommands()

    log.Println("Bot iniciado...")
    h.Bot.Start()
}

func (h *BotHandler) setupCommands() {
    h.Bot.Handle("/start", h.HandleStart)
    h.Bot.Handle("/registrar", h.HandleRegistrar)
    h.Bot.Handle("/ultimaalerta", h.HandleUltimaAlerta)
    h.Bot.Handle(tele.OnText, h.HandleText)
}

// Define command handlers (HandleStart, HandleRegistrar, HandleUltimaAlerta, HandleText)
func (h *BotHandler) HandleStart(c tele.Context) error {
	h.userStates[c.Chat().ID] = ""
	return c.Send("¡Bienvenido! Para registrar tu ESP32, usa uno de los siguientes comandos:\n\n" +
		"/registrar - Registrar un nuevo producto ESP32\n" +
		"/ultimaalerta - Ver la última alerta de tu sensor")
}

func (h *BotHandler) HandleRegistrar(c tele.Context) error {
	h.userStates[c.Chat().ID] = "waiting_serial"
	return c.Send("Por favor, ingresa el número de serial de tu ESP32:")
}

func (h *BotHandler) HandleUltimaAlertaV2(c tele.Context) error {
	chatID := c.Chat().ID
reading, err := h.ky026Service.GetLastReading(strconv.FormatInt(chatID, 10))
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
}

func (h *BotHandler) HandleText(c tele.Context) error {
	chatID := c.Chat().ID
	state := h.userStates[chatID]
	text := c.Text()

	switch state {
	case "waiting_serial":
		valid, err := h.esp32Service.ValidateAndLinkESP32(chatID, text)
		if err != nil {
			return c.Send("Error: " + err.Error())
		}
		if !valid {
			return c.Send("El número de serial no es válido. Por favor, verifica e intenta nuevamente.")
		}
		h.userStates[chatID] = ""
		return c.Send("¡ESP32 registrado exitosamente! Recibirás alertas cuando se detecte humo o fuego.")

	default:
		return c.Send("Por favor, usa /start para ver los comandos disponibles.")
	}
}
