package application

import (
	"errors"
	"telegramassist/internal/domain"
	"fmt"
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

// Añadir después de GetLastKY026ReadingByChat
// Add debugging to the ProcessAlert function
func (s *ESP32Service) ProcessAlert(alert *domain.Alert) ([]int64, error) {
    fmt.Printf("Processing alert for ESP32 with serial: %s\n", alert.NumeroSerie)
    
    // Verificar si el ESP32 existe
    esp32, err := s.repo.GetBySerial(alert.NumeroSerie)
    if err != nil {
        fmt.Printf("Error finding ESP32: %v\n", err)
        return nil, err
    }
    if esp32 == nil {
        fmt.Printf("ESP32 not found with serial: %s\n", alert.NumeroSerie)
        return nil, errors.New("ESP32 no encontrado")
    }

    fmt.Printf("Found ESP32: %+v\n", esp32)

    // Obtener los chats asociados al ESP32
    chatIDs, err := s.repo.GetChatsByESP32Serial(alert.NumeroSerie)
    if err != nil {
        fmt.Printf("Error getting chats: %v\n", err)
        return nil, err
    }

    fmt.Printf("Found %d chats for ESP32 %s: %v\n", len(chatIDs), alert.NumeroSerie, chatIDs)

    // Si es un sensor KY-026, guardar la lectura en la base de datos
    if alert.Sensor == "KY_026" {
        // Aquí podrías implementar la lógica para guardar la lectura
        fmt.Println("KY_026 sensor alert received")
    }

    return chatIDs, nil
}

// Add this method to the ESP32Service
func (s *ESP32Service) GetUserByESP32Serial(serial string) (*domain.User, error) {
	return s.repo.GetUserByESP32Serial(serial)
}
