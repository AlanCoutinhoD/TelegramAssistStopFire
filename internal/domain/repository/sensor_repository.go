package repository

import "telegramassist/internal/domain"

type KY026Repository interface {
    SaveReading(reading *domain.KY026Reading) error
    GetLastReading(serial string) (*domain.KY026Reading, error)
}