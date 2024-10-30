package services

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"net/http"
	"os"
)

func StartAmqpConnection() *amqp.Connection {
	connectionString := config.EnvAmqpConnection()
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		logger.Log.Error("Failed to connect to RabbitMQ")
		os.Exit(http.StatusInternalServerError)
	}

	return connection
}

func StartChannelConnection() *amqp.Channel {
	connection := StartAmqpConnection()
	channel, err := connection.Channel()
	if err != nil {
		logger.Log.Error("Failed to open a channel")
		os.Exit(http.StatusInternalServerError)
	}

	return channel
}

func SendDataToQueue(queue string, payload []byte) {
	channel := StartChannelConnection()

	q, internalError := channel.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if internalError != nil {
		logger.Log.Error(internalError.Error())

		os.Exit(http.StatusInternalServerError)
	}

	internalError = channel.PublishWithContext(context.Background(),
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		})

	if internalError != nil {
		logger.Log.Error(internalError.Error())

		os.Exit(http.StatusInternalServerError)
	}
}

func ConsumeQueue(queue string) <-chan amqp.Delivery {

	channel := StartChannelConnection()

	q, internalError := channel.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if internalError != nil {
		logger.Log.Error(internalError.Error())
		os.Exit(http.StatusInternalServerError)
	}

	msgs, internalError := channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if internalError != nil {
		logger.Log.Error(internalError.Error())
		os.Exit(http.StatusInternalServerError)
	}

	return msgs
}
