package usecase

import (
	"encoding/json"
	"nextclan/transaction-gateway/transaction-submit-service/internal/entity"
	"nextclan/transaction-gateway/transaction-submit-service/pkg/logger"
	"nextclan/transaction-gateway/transaction-submit-service/pkg/redis"

	"github.com/streadway/amqp"
)

//Use cases include:
/*
Given verified transaction When submit with validatedTransaction command Then the transaction should complete without errors.
Given verified transaction When receive transaction with verifiedTransaction command Then the transaction should publish into rabbitMQ without error.
*/

type ReceiveValidatedTransactionFromQueueUseCase struct {
	log        logger.Interface
	redisCache *redis.RedisCache
}

func NewReceiveValidatedTransaction(l logger.Interface, redis *redis.RedisCache) *ReceiveValidatedTransactionFromQueueUseCase {
	return &ReceiveValidatedTransactionFromQueueUseCase{
		log:        l,
		redisCache: redis,
	}
}

func (rvt *ReceiveValidatedTransactionFromQueueUseCase) Handle(d amqp.Delivery) {
	body := d.Body
	validatedTransaction := &entity.ValidatedTransaction{}
	err := json.Unmarshal(body, validatedTransaction)
	if err != nil {
		rvt.log.Debug("Problem parsing validated transaction : %v", err.Error())
		_ = d.Ack(false)
		panic(err)
	} else {
		rvt.log.Debug("%s Receive Message %s", d.ConsumerTag, validatedTransaction)
		val, err := rvt.redisCache.Get(validatedTransaction.TransactionId)

		rawTransaction := &entity.RawTransaction{}
		err = json.Unmarshal([]byte(val), rawTransaction)

		rvt.log.Debug("Deserialize %v", rawTransaction)

		response, err := LoaffinityClient.SendRawTransaction(rawTransaction.TransactionData, "0")
		if err != nil {
			rvt.log.Error("Problem parsing validated transaction : %v", err.Error())
		}
		if response.Error != nil {
			rvt.log.Error("RPC call failed: %d %s", response.Error.Code, response.Error.Message)
		}

		err = rvt.redisCache.Del(validatedTransaction.TransactionId)
		if err != nil {
			rvt.log.Error("Problem remove validated transaction from cache : %v", err.Error())
		}

		_ = d.Ack(true)
	}
}
