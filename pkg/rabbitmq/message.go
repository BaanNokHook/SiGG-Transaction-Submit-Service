package messaging

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// Defines our interface for connecting and consuming messages.
type IMessagingClient interface {
	Connect(connectionString string)
	Publish(msg []byte, exchangeName string, exchangeType string) error
	PublishOnQueue(msg []byte, exchangeName string, exchangeType string, queueName string) error
	Subscribe(exchangeName string, exchangeType string, consumerName string, handlerFunc func(amqp.Delivery)) error
	SubscribeToQueue(exchangeName string, exchangeType string, queueName string, consumerName string, handlerFunc func(amqp.Delivery)) error
	Close() error
}

// Real implementation, encapsulates a pointer to an amqp.Connection
type MessagingClient struct {
	conn *amqp.Connection
}

func (m *MessagingClient) Connect(connectionString string) {
	if connectionString == "" {
		panic("Cannot initialize connection to broker, connectionString not set. Have you initialized?")
	}

	var err error
	m.conn, err = amqp.Dial(fmt.Sprintf("%s/", connectionString))
	if err != nil {
		panic("Failed to connect to AMQP compatible broker at: " + connectionString)
	}
}

func (m *MessagingClient) Publish(body []byte, exchangeName string, exchangeType string) error {
	if m.conn == nil {
		panic("Tried to send message before connection was initialized. Don't do that.")
	}
	ch, err := m.conn.Channel() // Get a channel from the connection
	defer ch.Close()
	err = ch.ExchangeDeclare(
		exchangeName, // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	)
	failOnError(err, "Failed to register an Exchange")

	queue, err := ch.QueueDeclare( // Declare a queue that will be created if not exists with some args
		"",    // our queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	err = ch.QueueBind(
		queue.Name,   // name of the queue
		exchangeName, // bindingKey
		exchangeName, // sourceExchange
		false,        // noWait
		nil,          // arguments
	)

	err = ch.Publish( // Publishes a message onto the queue.
		exchangeName, // exchange
		exchangeName, // routing key      q.Name
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			Body: body, // Our JSON body as []byte
		})
	fmt.Printf("A message was sent: %v", body)
	return err
}

func (m *MessagingClient) PublishOnQueue(body []byte, exchangeName string, exchangeType string, queueName string) error {
	if m.conn == nil {
		panic("Tried to send message before connection was initialized. Don't do that.")
	}
	ch, err := m.conn.Channel() // Get a channel from the connection
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName, // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	)
	failOnError(err, "Failed to register an Exchange")

	queue, err := ch.QueueDeclare( // Declare a queue that will be created if not exists with some args
		queueName, // our queue name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	err = ch.QueueBind(
		queue.Name,   // name of the queue
		queue.Name,   // bindingKey
		exchangeName, // sourceExchange
		false,        // noWait
		nil,          // arguments
	)

	// Publishes a message onto the queue.
	err = ch.Publish(
		exchangeName, // exchange
		queue.Name,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body, // Our JSON body as []byte
		})
	fmt.Printf("A message was sent to queue %v: %v", queueName, body)
	return err
}

func (m *MessagingClient) Subscribe(exchangeName string, exchangeType string, consumerName string, handlerFunc func(amqp.Delivery)) error {
	ch, err := m.conn.Channel()
	failOnError(err, "Failed to open a channel")
	// defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName, // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	)
	failOnError(err, "Failed to register an Exchange")

	log.Printf("declared Exchange, declaring Queue (%s)", "")
	queue, err := ch.QueueDeclare(
		"",    // name of the queue
		false, // durable
		false, // delete when usused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	failOnError(err, "Failed to register an Queue")

	log.Printf("declared Queue (%d messages, %d consumers), binding to Exchange (key '%s')",
		queue.Messages, queue.Consumers, exchangeName)

	err = ch.QueueBind(
		queue.Name,   // name of the queue
		exchangeName, // bindingKey
		exchangeName, // sourceExchange
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Bind: %s", err)
	}

	msgs, err := ch.Consume(
		queue.Name,   // queue
		consumerName, // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	failOnError(err, "Failed to register a consumer")

	go consumeLoop(msgs, handlerFunc)
	return nil
}

func (m *MessagingClient) SubscribeToQueue(exchangeName string, exchangeType string, queueName string, consumerName string, handlerFunc func(amqp.Delivery)) error {
	ch, err := m.conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		exchangeName, // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	)
	failOnError(err, "Failed to register an Exchange")

	log.Printf("Declaring Queue (%s)", queueName)
	queue, err := ch.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	failOnError(err, "Failed to register an Queue")

	err = ch.QueueBind(
		queue.Name,   // name of the queue
		queue.Name,   // bindingKey
		exchangeName, // sourceExchange
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Bind: %s", err)
	}

	msgs, err := ch.Consume(
		queue.Name,   // queue
		consumerName, // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	failOnError(err, "Failed to register a consumer")

	go consumeLoop(msgs, handlerFunc)
	return nil
}

func (m *MessagingClient) Close() error {
	if m.conn != nil {
		err := m.conn.Close()
		return err
	}
	return nil
}

func consumeLoop(deliveries <-chan amqp.Delivery, handlerFunc func(d amqp.Delivery)) {
	for d := range deliveries {
		// Invoke the handlerFunc func we passed as parameter.
		handlerFunc(d)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
