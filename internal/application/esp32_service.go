package application

import (
	"errors"
	"telegramassist/internal/domain"
	
)

type ESP32Service struct {
	repo domain.ESP32Repository
	ky026Service *KY026Service
}

func NewESP32Service(repo domain.ESP32Repository, ky026Service *KY026Service) *ESP32Service {
	return &ESP32Service{
		repo: repo,
		ky026Service: ky026Service,
	}
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
    return s.ky026Service.GetLastReading(serial)
}

func (s *ESP32Service) ProcessAlert(alert *domain.Alert) ([]int64, error) {
    if alert.Sensor == "KY_026" {
        if err := s.ky026Service.ProcessKY026Alert(alert); err != nil {
            return nil, err
        }
    }

    chatIDs, err := s.repo.GetChatsByESP32Serial(alert.NumeroSerie)
    if err != nil {
        return nil, err
    }

    return chatIDs, nil
}

// this method to the ESP32Service
func (s *ESP32Service) GetUserByESP32Serial(serial string) (*domain.User, error) {
	return s.repo.GetUserByESP32Serial(serial)
}
