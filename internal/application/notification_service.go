package application

import (
    "telegramassist/internal/domain"
    tele "gopkg.in/telebot.v3"
    "fmt"
)

type NotificationService struct {
    bot *tele.Bot
}

func NewNotificationService(bot *tele.Bot) *NotificationService {
    return &NotificationService{bot: bot}
}

func (s *NotificationService) SendTelegramNotification(chatID int64, alert *domain.Alert) error {
    estadoTexto := "Desactivado"
    if alert.Estado == 1 {
        estadoTexto = "Activado"
    }

    mensaje := fmt.Sprintf("🚨 *ALERTA DE SENSOR* 🚨\n\nSensor: %s\nEstado: %s\nActivación: %s\nDesactivación: %s",
        alert.Sensor, estadoTexto, alert.FechaActivacion, alert.FechaDesactivacion)

    _, err := s.bot.Send(&tele.Chat{ID: chatID}, mensaje)
    return err
}