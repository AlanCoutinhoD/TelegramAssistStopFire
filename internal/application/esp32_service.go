package application

import (
	"errors"
	"telegramassist/internal/domain"
)

type ESP32Service struct {
	repo domain.ESP32Repository
}

func NewESP32Service(repo domain.ESP32Repository) *ESP32Service {
	return &ESP32Service{repo: repo}
}

func (s *ESP32Service) ValidateAndLinkESP32(chatID int64, serial string) (bool, error) {
	esp32, err := s.repo.GetBySerial(serial)
	if err != nil {
		return false, err
	}
	if esp32 == nil {
		return false, errors.New("ESP32 no encontrado")
	}

	err = s.repo.LinkChatToESP32(chatID, serial)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *ESP32Service) GetLastKY026Reading(serial string) (*domain.KY026Reading, error) {
	return s.repo.GetLastKY026Reading(serial)
}

func (s *ESP32Service) GetLastKY026ReadingByChat(chatID int64) (*domain.KY026Reading, error) {
	serial, err := s.repo.GetESP32SerialByChat(chatID)
	if err != nil {
		return nil, err
	}
	if serial == "" {
		return nil, errors.New("No hay un ESP32 registrado para este chat")
	}
	return s.repo.GetLastKY026Reading(serial)
}
