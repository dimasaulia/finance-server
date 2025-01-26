package transaction_validation

type NewTransactionRequest struct {
	IdUser               int64   `json:"id_user" validate:"required,number"`
	IdAccount            int64   `json:"id_account" validate:"required,number"`
	TransactionType      string  `json:"transaction_type" validate:"required,alphaspace"`
	TransactionGroup     string  `json:"transaction_group" validate:"required,alphaspace"`
	IdAccountDestination *int64  `json:"id_transaction_amount" validate:"omitempty,number"`
	Amount               float64 `json:"amount" validate:"required,number,gte=0"`
}

type TransactionResponse struct {
	IdTransaction int64   `json:"id_transaction"`
	Amount        float64 `json:"amount"`
}
