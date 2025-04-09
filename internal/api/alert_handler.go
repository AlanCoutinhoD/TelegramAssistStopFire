package api

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "telegramassist/internal/application"
    "telegramassist/internal/domain"
    "telegramassist/internal/infrastructure/rabbitmq"
)

type UserNotification struct {
    UserID        int    `json:"user_id"`
    Username      string `json:"username"`
    Email         string `json:"email"`
    SensorType    string `json:"sensor_type"`
    Estado        int    `json:"estado"`
    Activacion    string `json:"activacion"`
    Desactivacion string `json:"desactivacion"`
    NumeroSerie   string `json:"numero_serie"`
}

type AlertHandler struct {
    esp32Service        *application.ESP32Service
    notificationService *application.NotificationService
    rabbitMQService     *rabbitmq.RabbitMQService
}

func NewAlertHandler(
    esp32Service *application.ESP32Service,
    notificationService *application.NotificationService,
    rabbitMQService *rabbitmq.RabbitMQService,
) *AlertHandler {
    return &AlertHandler{
        esp32Service:        esp32Service,
        notificationService: notificationService,
        rabbitMQService:     rabbitMQService,
    }
}

func (h *AlertHandler) parseAlert(r *http.Request) (*domain.Alert, error) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading request body: %v", err)
    }
    
    r.Body = io.NopCloser(bytes.NewBuffer(body))
    
    var alert domain.Alert
    if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
        return nil, fmt.Errorf("error decoding alert: %v", err)
    }
    
    return &alert, nil
}

func (h *AlertHandler) createUserNotification(user *domain.User, alert *domain.Alert) UserNotification {
    return UserNotification{
        UserID:        user.ID,
        Username:      user.Username,
        Email:         user.Email,
        SensorType:    alert.Sensor,
        Estado:        alert.Estado,
        Activacion:    alert.FechaActivacion,
        Desactivacion: alert.FechaDesactivacion,
        NumeroSerie:   alert.NumeroSerie,
    }
}

func (h *AlertHandler) sendSuccessResponse(w http.ResponseWriter, chatCount int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status":        "success",
        "message":       "Alerta procesada correctamente",
        "chats_notified": fmt.Sprintf("%d", chatCount),
    })
}

func (h *AlertHandler) HandleAlert(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    alert, err := h.parseAlert(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    chatIDs, err := h.processAlert(alert)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    h.sendSuccessResponse(w, len(chatIDs))
}

func (h *AlertHandler) processAlert(alert *domain.Alert) ([]int64, error) {
    user, err := h.esp32Service.GetUserByESP32Serial(alert.NumeroSerie)
    if err != nil {
        return nil, fmt.Errorf("error getting user: %v", err)
    }

    if user != nil {
        notification := h.createUserNotification(user, alert)
        if err := h.rabbitMQService.PublishNotification(notification); err != nil {
            fmt.Printf("Error sending to RabbitMQ: %v\n", err)
        }
    }

    chatIDs, err := h.esp32Service.ProcessAlert(alert)
    if err != nil {
        return nil, fmt.Errorf("error processing alert: %v", err)
    }

    for _, chatID := range chatIDs {
        if err := h.notificationService.SendTelegramNotification(chatID, alert); err != nil {
            fmt.Printf("Error sending telegram notification to %d: %v\n", chatID, err)
        }
    }

    return chatIDs, nil
}