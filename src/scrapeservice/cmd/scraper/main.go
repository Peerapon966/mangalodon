package main

import (
	"context"
	"errors"
	"log"
	"os"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

var brokerURI = os.Getenv("RABBITMQ_URL")
var queueName = os.Getenv("RABBITMQ_QUEUE")

func main() {
	log.Printf("Broker URI: %v", brokerURI)

	ctx := context.Background()
	env := rmq.NewEnvironment(brokerURI, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		log.Panicf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		_ = env.CloseConnections(context.Background())
	}()

	_, err = conn.Management().DeclareQueue(ctx, &rmq.QuorumQueueSpecification{Name: queueName})
	if err != nil {
		log.Panicf("Failed to declare a queue: %v", err)
	}

	consumer, err := conn.NewConsumer(ctx, queueName, nil)
	if err != nil {
		log.Panicf("Failed to create consumer: %v", err)
	}
	defer func() { _ = consumer.Close(context.Background()) }()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	// for {
	delivery, err := consumer.Receive(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		log.Panicf("Failed to receive a message: %v", err)
	}
	msg := delivery.Message()
	var body string
	if len(msg.Data) > 0 {
		body = string(msg.Data[0])
	}
	log.Printf("Received a message: %s", body)
	err = delivery.Accept(ctx)
	if err != nil {
		log.Panicf("Failed to accept message: %v", err)
	}
	// }
}
