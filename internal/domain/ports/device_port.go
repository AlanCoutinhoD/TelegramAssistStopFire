package ports

import "telegramassist/internal/domain"

type DeviceManager interface {
    GetDevice(serial string) (*domain.ESP32, error)
    GetLinkedChats(serial string) ([]int64, error)
    LinkDeviceToChat(chatID int64, serial string) error
}