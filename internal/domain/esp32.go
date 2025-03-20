package domain

type ESP32 struct {
	ID     int
	Serial string
}

type TelegramChat struct {
	ID          int
	ChatID      int64
	ESP32Serial string
}

type KY026Reading struct {
	ID              int
	ESP32Serial     string
	FechaActivacion string
	Estado          string
}

type ESP32Repository interface {
	GetBySerial(serial string) (*ESP32, error)
	LinkChatToESP32(chatID int64, serial string) error
	GetLastKY026Reading(serial string) (*KY026Reading, error)
	GetESP32SerialByChat(chatID int64) (string, error)
}
