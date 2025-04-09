package rabbitmq

import (
    "encoding/json"
    "fmt"
    "os"
    "github.com/streadway/amqp"
)

type RabbitMQService struct {
    url       string
    queueName string
}

func NewRabbitMQService() *RabbitMQService {
    return &RabbitMQService{
        url:       os.Getenv("RABBITMQ_URL"),
        queueName: os.Getenv("RABBITMQ_QUEUE"),
    }
}

func (s *RabbitMQService) PublishNotification(notification interface{}) error {
    conn, err := amqp.Dial(s.url)
    if err != nil {
        return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        return fmt.Errorf("failed to open channel: %v", err)
    }
    defer ch.Close()

    body, err := json.Marshal(notification)
    if err != nil {
        return fmt.Errorf("failed to marshal notification: %v", err)
    }

    err = ch.Publish("", s.queueName, false, false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        })
    return err
}