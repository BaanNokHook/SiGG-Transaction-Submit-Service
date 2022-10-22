package entity

type RawTransaction struct {
	TransactionData string `json: "transactionData"`
	MaxGasFee       string `json: "maxGasFee"`
}
