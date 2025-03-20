package mysql

import (
	"database/sql"
	"fmt"
	"os"
	"telegramassist/internal/domain"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository() (*MySQLRepository, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &MySQLRepository{db: db}, nil
}

func (r *MySQLRepository) GetBySerial(serial string) (*domain.ESP32, error) {
	esp := &domain.ESP32{}
	err := r.db.QueryRow("SELECT id, serial FROM esp32m WHERE serial = ?", serial).
		Scan(&esp.ID, &esp.Serial)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return esp, err
}

func (r *MySQLRepository) LinkChatToESP32(chatID int64, serial string) error {
	_, err := r.db.Exec(
		"INSERT INTO telegram_chats (chat_id, esp32_serial) VALUES (?, ?)",
		chatID, serial)
	return err
}

func (r *MySQLRepository) GetLastKY026Reading(serial string) (*domain.KY026Reading, error) {
	reading := &domain.KY026Reading{}
	err := r.db.QueryRow(`
		SELECT idKY_026, esp32_serial, fecha_activacion, estado 
		FROM ky_026 
		WHERE esp32_serial = ? 
		ORDER BY idKY_026 DESC 
		LIMIT 1`, serial).
		Scan(&reading.ID, &reading.ESP32Serial, &reading.FechaActivacion, &reading.Estado)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return reading, err
}

func (r *MySQLRepository) GetESP32SerialByChat(chatID int64) (string, error) {
	var serial string
	err := r.db.QueryRow("SELECT esp32_serial FROM telegram_chats WHERE chat_id = ? ORDER BY created_at DESC LIMIT 1", chatID).
		Scan(&serial)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return serial, err
}
