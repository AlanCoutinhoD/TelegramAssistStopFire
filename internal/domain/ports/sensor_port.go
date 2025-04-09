package ports

import "telegramassist/internal/domain"

type KY026Manager interface {
    GetLastReading(serial string) (*domain.KY026Reading, error)
    SaveReading(reading *domain.KY026Reading) error
    ProcessAlert(alert *domain.Alert) error
}