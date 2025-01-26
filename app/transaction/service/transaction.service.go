package transaction_service

import v "finance/app/transaction/validation"

type ITransactionService interface {
	CreateNewTransaction(req *v.NewTransactionRequest) (*v.TransactionResponse, error)
}
