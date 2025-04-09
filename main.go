package main

import (
    "log"
    "telegramassist/internal/api"
    "telegramassist/internal/application"
    "telegramassist/internal/infrastructure/mysql"
    "telegramassist/internal/infrastructure/rabbitmq"
    "telegramassist/internal/bot"
    "telegramassist/internal/server"
    

    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    // Initialize MySQL Repository
    mysqlRepo, err := mysql.NewMySQLRepository()
    if err != nil {
        log.Fatal(err)
    }

    // Initialize Services
    ky026Service := application.NewKY026Service(mysqlRepo)
    esp32Service := application.NewESP32Service(mysqlRepo, ky026Service)
    
    // Initialize Bot Handler
    botHandler := bot.NewBotHandler(esp32Service, ky026Service)
    botHandler.Start()

    // Initialize RabbitMQ Service
    rabbitMQService := rabbitmq.NewRabbitMQService()

    // Initialize Notification Service with the bot
    notificationService := application.NewNotificationService(botHandler.Bot)

    // Initialize Alert Handler with correct services
    alertHandler := api.NewAlertHandler(
        esp32Service,
        notificationService,
        rabbitMQService,
    )

    // Initialize and start the HTTP server
    server.StartHTTPServer(alertHandler)
}