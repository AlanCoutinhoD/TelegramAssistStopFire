package mysql

import (
	"database/sql"
	"fmt"
	"os"
	"telegramassist/internal/domain"
	
	"strconv" 

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
	err := r.db.QueryRow("SELECT idESP32, numero_serie FROM ESP32 WHERE numero_serie = ?", serial).
		Scan(&esp.ID, &esp.Serial) 
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return esp, err
}

// Update the LinkChatToESP32 method
func (r *MySQLRepository) LinkChatToESP32(chatID int64, serial string) error {
	_, err := r.db.Exec(
		"INSERT INTO telegram_chats (chat_id, esp32_serial) VALUES (?, ?)",
		chatID, serial)
	return err
}

// Update the GetLastKY026Reading method
func (r *MySQLRepository) GetLastKY026Reading(serial string) (*domain.KY026Reading, error) {
	reading := &domain.KY026Reading{}
	err := r.db.QueryRow(`
		SELECT idKY_026, numero_serie, fecha_activacion, estado 
		FROM KY_026 
		WHERE numero_serie = ? 
		ORDER BY idKY_026 DESC 
		LIMIT 1`, serial).
		Scan(&reading.ID, &reading.ESP32Serial, &reading.FechaActivacion, &reading.Estado)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return reading, err
}

// Update the GetESP32SerialByChat method
func (r *MySQLRepository) GetESP32SerialByChat(chatID int64) (string, error) {
	var serial string
	err := r.db.QueryRow("SELECT esp32_serial FROM telegram_chats WHERE chat_id = ? ORDER BY created_at DESC LIMIT 1", chatID).
		Scan(&serial)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return serial, err
}


// Add debugging to the GetChatsByESP32Serial method
func (r *MySQLRepository) GetChatsByESP32Serial(serial string) ([]int64, error) {
    fmt.Printf("Querying chats for ESP32 with serial: %s\n", serial)
    
    query := "SELECT chat_id FROM telegram_chats WHERE esp32_serial = ?"
    fmt.Printf("Executing query: %s with parameter: %s\n", query, serial)
    
    rows, err := r.db.Query(query, serial)
    if err != nil {
        fmt.Printf("Error executing query: %v\n", err)
        return nil, err
    }
    defer rows.Close()

    var chatIDs []int64
    for rows.Next() {
        var chatID int64
        if err := rows.Scan(&chatID); err != nil {
            fmt.Printf("Error scanning row: %v\n", err)
            return nil, err
        }
        chatIDs = append(chatIDs, chatID)
        fmt.Printf("Found chat ID: %d\n", chatID)
    }

    if err := rows.Err(); err != nil {
        fmt.Printf("Error after scanning rows: %v\n", err)
        return nil, err
    }

    fmt.Printf("Total chats found: %d\n", len(chatIDs))
    return chatIDs, nil
}

// Add this method to the MySQLRepository
// Update the GetUserByESP32Serial method to handle NULL values
func (r *MySQLRepository) GetUserByESP32Serial(serial string) (*domain.User, error) {
	var userID sql.NullInt64
	
	// First, get the user ID from the ESP32 table, handling NULL values
	err := r.db.QueryRow("SELECT idUser FROM ESP32 WHERE numero_serie = ?", serial).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No ESP32 found
		}
		return nil, err
	}
	
	// Check if userID is NULL or not valid
	if !userID.Valid {
		return nil, nil // No user associated with this ESP32
	}
	
	// Then, get the user details
	user := &domain.User{}
	err = r.db.QueryRow("SELECT id, username, email FROM users WHERE id = ?", userID.Int64).
		Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No user found
		}
		return nil, err
	}
	
	return user, nil
}

// Implement KY026Manager interface
func (r *MySQLRepository) GetLastReading(serial string) (*domain.KY026Reading, error) {
    reading := &domain.KY026Reading{}
    err := r.db.QueryRow(`
        SELECT idKY_026, numero_serie, fecha_activacion, estado 
        FROM KY_026 
        WHERE numero_serie = ? 
        ORDER BY idKY_026 DESC 
        LIMIT 1`, serial).
        Scan(&reading.ID, &reading.ESP32Serial, &reading.FechaActivacion, &reading.Estado)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return reading, err
}

func (r *MySQLRepository) SaveReading(reading *domain.KY026Reading) error {
    _, err := r.db.Exec(
        "INSERT INTO KY_026 (numero_serie, fecha_activacion, estado) VALUES (?, ?, ?)",
        reading.ESP32Serial, reading.FechaActivacion, reading.Estado)
    return err
}

func (r *MySQLRepository) ProcessAlert(alert *domain.Alert) error {
    if alert.Sensor != "KY_026" {
        return nil
    }
    reading := &domain.KY026Reading{
        ESP32Serial:     alert.NumeroSerie,
        FechaActivacion: alert.FechaActivacion,
        Estado:          strconv.Itoa(alert.Estado),
    }
    return r.SaveReading(reading)
}

// Add these methods to implement NotificationManager interface
func (r *MySQLRepository) NotifyUsers(chatIDs []int64, alert *domain.Alert) error {
    // Implement notification logic here
    return nil
}

func (r *MySQLRepository) GetLinkedChats(serial string) ([]int64, error) {
    return r.GetChatsByESP32Serial(serial)
}

// Add this method to implement DeviceManager interface
func (r *MySQLRepository) GetDevice(serial string) (*domain.ESP32, error) {
    return r.GetBySerial(serial)
}

// Add these methods to implement missing interfaces
func (r *MySQLRepository) LinkDeviceToChat(chatID int64, serial string) error {
    return r.LinkChatToESP32(chatID, serial)
}

func (r *MySQLRepository) GetNotificationPreferences(userID int) (domain.NotificationPreferences, error) {
    // Default preferences for now
    return domain.NotificationPreferences{
        EnableTelegram: true,
        EnableEmail:    false,
        EnableSMS:      false,
    }, nil
}
