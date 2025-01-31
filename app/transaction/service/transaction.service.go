package transaction_service

import (
	v "finance/app/transaction/validation"
	u "finance/utility/response"
)

type ITransactionService interface {
	CreateNewTransaction(req *v.NewTransactionRequest) (*[]v.TransactionResponse, error)
	UpdateTransaction(req *v.UpdateTransactionRequest) (*[]v.TransactionResponse, error)
	DeleteTransaction(req *v.DeleteTransactionRequest) error
	GetUserTransaction(req *u.StandarGetRequest, data *v.UserTransactionDetailRequest) (*[]v.TransactionData, error)
	CreateNewSubTransaction(req *v.NewSubTransactionRequest) (*[]v.TransactionResponse, error)
}
