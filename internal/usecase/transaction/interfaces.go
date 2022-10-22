package usecase

import (
	"nextclan/transaction-gateway/transaction-submit-service/pkg/loaffinity"
	messaging "nextclan/transaction-gateway/transaction-submit-service/pkg/rabbitmq"

	"github.com/streadway/amqp"
)

type (
	ReceiveValidatedTransactionFromQueue interface {
		Handle(d amqp.Delivery)
	}
)

var MessagingClient messaging.IMessagingClient
var LoaffinityClient loaffinity.ILoaffinityClient
