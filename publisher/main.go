package main

import (
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	conn := waitForRabbitMQ(amqpURL, 10, 3*time.Second)
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"tasks", // name
		true,    // durable
		false,   // auto-delete
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	for i := 1; i <= 5; i++ {
		body := fmt.Sprintf("Task #%d: Send email to user%d@example.com", i, i)
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		if err != nil {
			log.Printf("Failed to publish: %v", err)
		} else {
			log.Printf("Published: %s", body)
		}
		time.Sleep(1 * time.Second)
	}
}

func waitForRabbitMQ(amqpURL string, retries int, delay time.Duration) *amqp.Connection {
	for i := 1; i <= retries; i++ {
		conn, err := amqp.Dial(amqpURL)
		if err == nil {
			return conn
		}
		log.Printf("RabbitMQ not ready (attempt %d/%d): %v", i, retries, err)
		time.Sleep(delay)
	}
	log.Fatal("RabbitMQ not available after retries")
	return nil
}
