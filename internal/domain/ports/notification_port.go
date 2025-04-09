package ports

import "telegramassist/internal/domain"

type NotificationManager interface {
    NotifyUsers(chatIDs []int64, alert *domain.Alert) error
    GetNotificationPreferences(userID int) (NotificationPreferences, error)
}

type NotificationPreferences struct {
    EnableTelegram bool
    EnableEmail    bool
    EnableSMS     bool
}