package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"telegramassist/internal/application"
	"telegramassist/internal/domain"

	"github.com/streadway/amqp" // Add RabbitMQ library
	tele "gopkg.in/telebot.v3"
)

type AlertHandler struct {
	esp32Service *application.ESP32Service
	bot          *tele.Bot
}

// UserNotification represents the data to be sent to the queue
type UserNotification struct {
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	SensorType   string `json:"sensor_type"`
	Estado       int    `json:"estado"`
	Activacion   string `json:"activacion"`
	Desactivacion string `json:"desactivacion"`
	NumeroSerie  string `json:"numero_serie"`
}

func NewAlertHandler(esp32Service *application.ESP32Service, bot *tele.Bot) *AlertHandler {
	return &AlertHandler{
		esp32Service: esp32Service,
		bot:          bot,
	}
}

// HandleAlert processes incoming alert requests and sends notifications to registered chats
func (h *AlertHandler) HandleAlert(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    // Log request body for debugging
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error al leer el cuerpo de la solicitud: "+err.Error(), http.StatusBadRequest)
        return
    }
    
    // Log the received payload
    fmt.Printf("Received alert payload: %s\n", string(body))
    
    // Restore the body for further processing
    r.Body = io.NopCloser(bytes.NewBuffer(body))

    var alert domain.Alert
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&alert); err != nil {
        http.Error(w, "Error al decodificar el cuerpo de la solicitud: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Log the parsed alert
    fmt.Printf("Parsed alert: %+v\n", alert)

    // Get user information associated with the ESP32
    user, err := h.esp32Service.GetUserByESP32Serial(alert.NumeroSerie)
    if err != nil {
        fmt.Printf("Error getting user info: %v\n", err)
        http.Error(w, "Error al obtener información del usuario: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // If user found, send notification to RabbitMQ
    if user != nil {
        // Create notification object
        notification := UserNotification{
            UserID:        user.ID,
            Username:      user.Username,
            Email:         user.Email,
            SensorType:    alert.Sensor,
            Estado:        alert.Estado,
            Activacion:    alert.FechaActivacion,
            Desactivacion: alert.FechaDesactivacion,
            NumeroSerie:   alert.NumeroSerie,
        }

        // Send to RabbitMQ
        err = sendToRabbitMQ(notification)
        if err != nil {
            fmt.Printf("Error sending to RabbitMQ: %v\n", err)
        } else {
            fmt.Printf("Notification sent to RabbitMQ for user %s\n", user.Username)
        }
    } else {
        fmt.Printf("No user found for ESP32 with serial %s\n", alert.NumeroSerie)
    }

    // Procesar la alerta y obtener los chats a notificar
    chatIDs, err := h.esp32Service.ProcessAlert(&alert)
    if err != nil {
        fmt.Printf("Error processing alert: %v\n", err)
        http.Error(w, "Error al procesar la alerta: "+err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Printf("Found %d chats to notify\n", len(chatIDs))

    // Enviar notificaciones a los chats
    for _, chatID := range chatIDs {
        // Determinar el mensaje según el estado
        estadoTexto := "Desactivado"
        if alert.Estado == 1 {
            estadoTexto = "Activado"
        }

        // Fix the markdown formatting to ensure proper escaping
        mensaje := "🚨 *ALERTA DE SENSOR* 🚨\n\n" +
            "Sensor: " + alert.Sensor + "\n" +
            "Estado: " + estadoTexto + "\n" +
            "Activación: " + alert.FechaActivacion + "\n" +
            "Desactivación: " + alert.FechaDesactivacion

        // Log the message being sent
        fmt.Printf("Sending message to chat %d: %s\n", chatID, mensaje)

        // Enviar mensaje al chat - removed Markdown parsing
        _, err := h.bot.Send(&tele.Chat{ID: chatID}, mensaje)
        
        if err != nil {
            fmt.Printf("Error sending message to chat %d: %v\n", chatID, err)
        } else {
            fmt.Printf("Message sent successfully to chat %d\n", chatID)
        }
    }

    // Responder con éxito
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "success",
        "message": "Alerta procesada correctamente",
        "chats_notified": fmt.Sprintf("%d", len(chatIDs)),
    })
}

// sendToRabbitMQ sends a notification to RabbitMQ
func sendToRabbitMQ(notification UserNotification) error {
    // Get RabbitMQ connection details from environment variables
    rabbitURL := os.Getenv("RABBITMQ_URL")
    if rabbitURL == "" {
        rabbitURL = "amqp://guest:guest@localhost:5672/" // Default if not set
    }
    
    queueName := os.Getenv("RABBITMQ_QUEUE")
    if queueName == "" {
        queueName = "userNotification" // Default queue name
    }

    fmt.Printf("Connecting to RabbitMQ at: %s\n", rabbitURL)
    
    // Connect to RabbitMQ
    conn, err := amqp.Dial(rabbitURL)
    if err != nil {
        fmt.Printf("Failed to connect to RabbitMQ: %v\n", err)
        return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
    }
    fmt.Println("Successfully connected to RabbitMQ")
    defer conn.Close()

    // Create a channel
    ch, err := conn.Channel()
    if err != nil {
        fmt.Printf("Failed to open a channel: %v\n", err)
        return fmt.Errorf("failed to open a channel: %v", err)
    }
    fmt.Println("Successfully opened channel")
    defer ch.Close()

    // Declare the queue (creates it if it doesn't exist)
    q, err := ch.QueueDeclare(
        queueName, // name
        true,      // durable
        false,     // delete when unused
        false,     // exclusive
        false,     // no-wait
        nil,       // arguments
    )
    if err != nil {
        fmt.Printf("Failed to declare queue: %v\n", err)
        return fmt.Errorf("failed to declare a queue: %v", err)
    }
    fmt.Printf("Successfully declared queue: %s\n", queueName)

    // Convert notification to JSON
    body, err := json.Marshal(notification)
    if err != nil {
        fmt.Printf("Failed to marshal notification: %v\n", err)
        return fmt.Errorf("failed to marshal notification: %v", err)
    }
    
    fmt.Printf("Notification JSON: %s\n", string(body))

    // Publish message to queue
    err = ch.Publish(
        "",     // exchange
        q.Name, // routing key
        false,  // mandatory
        false,  // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        })
    if err != nil {
        fmt.Printf("Failed to publish message: %v\n", err)
        return fmt.Errorf("failed to publish a message: %v", err)
    }

    fmt.Printf("Successfully sent notification to queue %s\n", queueName)
    return nil
}