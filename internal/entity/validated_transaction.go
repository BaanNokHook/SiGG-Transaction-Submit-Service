package entity

type ValidatedTransaction struct {
	TransactionId string `json: "transactionId"`
	Signature     string `json: "signature"`
}
