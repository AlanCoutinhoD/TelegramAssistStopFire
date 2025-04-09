package application

import (
    "telegramassist/internal/domain"
    "telegramassist/internal/domain/ports"
)

type KY026Service struct {
    sensorManager ports.KY026Manager
}

func (s *KY026Service) GetLastReading(serial string) (*domain.KY026Reading, error) {
    return s.sensorManager.GetLastReading(serial)
}

func (s *KY026Service) ProcessKY026Alert(alert *domain.Alert) error {
    return s.sensorManager.ProcessAlert(alert)
}

func NewKY026Service(sensorManager ports.KY026Manager) *KY026Service {
    return &KY026Service{
        sensorManager: sensorManager,
    }
}