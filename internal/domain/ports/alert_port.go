package ports

import "telegramassist/internal/domain"

type AlertNotifier interface {
    NotifyAlert(alert *domain.Alert) error
}