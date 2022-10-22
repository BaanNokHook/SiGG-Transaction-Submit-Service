package loaffinity

import (
	"nextclan/transaction-gateway/transaction-submit-service/pkg/logger"

	"github.com/KeisukeYamashita/go-jsonrpc"
)

type ILoaffinityClient interface {
	SendRawTransaction(transactionData string, maxGasFee string) (*jsonrpc.RPCResponse, error)
}

type LoaffinityClient struct {
	rpcClient *jsonrpc.RPCClient
	log       logger.Interface
}

func NewLoaffinityClient(url string, user string, password string, log logger.Interface) *LoaffinityClient {
	rpcClient := jsonrpc.NewRPCClient(url)
	rpcClient.SetBasicAuth(user, password)
	return &LoaffinityClient{
		rpcClient: rpcClient,
		log:       log,
	}
}

func (l *LoaffinityClient) SendRawTransaction(transactionData string, maxGasFee string) (*jsonrpc.RPCResponse, error) {
	response, err := l.rpcClient.Call("sendrawtransaction", transactionData, maxGasFee)
	if err != nil {
		l.log.Debug("RPC call failed: %v", err.Error())
		return nil, err
	}
	return response, err
}
